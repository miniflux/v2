// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/model"
)

func newCSRFTestSession(authenticated bool) *model.WebSession {
	session, _ := model.NewWebSession("test", "127.0.0.1")
	if authenticated {
		session.SetUser(&model.User{ID: 1})
	}
	return session
}

func withCSRFTestSession(r *http.Request, session *model.WebSession) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), request.WebSessionContextKey, session))
}

func newMultipartBody(t *testing.T, fields map[string]string, fileContent []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if fileContent != nil {
		part, err := writer.CreateFormFile("file", "feeds.opml")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(fileContent); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}

func TestCSRFMiddlewareIgnoresMultipartBodyWithoutAuthentication(t *testing.T) {
	session := newCSRFTestSession(false)
	body, contentType := newMultipartBody(t, map[string]string{"csrf": session.CSRF()}, []byte("<opml/>"))

	r := httptest.NewRequest(http.MethodPost, "/login", body)
	r.Header.Set("Content-Type", contentType)
	r = withCSRFTestSession(r, session)

	handler := newCSRFMiddleware("/base").handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal(`The next handler should not be called`)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Fatalf(`Unexpected status code, got: %d`, w.Code)
	}

	if location := w.Header().Get("Location"); location != "/base/" {
		t.Fatalf(`Unexpected redirect location, got: %q`, location)
	}

	if r.MultipartForm != nil {
		t.Fatal(`The multipart form should not be parsed before authentication`)
	}
}

func TestCSRFMiddlewareAcceptsURLEncodedFormWithoutAuthentication(t *testing.T) {
	session := newCSRFTestSession(false)
	form := url.Values{"csrf": {session.CSRF()}, "username": {"admin"}}

	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = withCSRFTestSession(r, session)

	called := false
	handler := newCSRFMiddleware("").handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if username := r.FormValue("username"); username != "admin" {
			t.Fatalf(`Unexpected username, got: %q`, username)
		}
	}))

	handler.ServeHTTP(httptest.NewRecorder(), r)

	if !called {
		t.Fatal(`The next handler should be called`)
	}
}

func TestCSRFMiddlewareAcceptsHeaderWithoutParsingBody(t *testing.T) {
	session := newCSRFTestSession(false)

	r := httptest.NewRequest(http.MethodPost, "/webauthn/login/finish", strings.NewReader(`{}`))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("X-Csrf-Token", session.CSRF())
	r = withCSRFTestSession(r, session)

	called := false
	handler := newCSRFMiddleware("").handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	handler.ServeHTTP(httptest.NewRecorder(), r)

	if !called {
		t.Fatal(`The next handler should be called`)
	}

	if r.Form != nil {
		t.Fatal(`The request body should not be parsed when the header token is valid`)
	}
}

func TestCSRFMiddlewareAcceptsMultipartFormWhenAuthenticated(t *testing.T) {
	session := newCSRFTestSession(true)
	body, contentType := newMultipartBody(t, map[string]string{"csrf": session.CSRF()}, []byte("<opml/>"))

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", contentType)
	r = withCSRFTestSession(r, session)

	called := false
	handler := newCSRFMiddleware("").handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf(`Unable to read the uploaded file: %v`, err)
		}
		file.Close()
	}))

	handler.ServeHTTP(httptest.NewRecorder(), r)

	if !called {
		t.Fatal(`The next handler should be called`)
	}
}

func TestCSRFMiddlewareRejectsQueryStringToken(t *testing.T) {
	session := newCSRFTestSession(true)

	r := httptest.NewRequest(http.MethodPost, "/feeds/refresh?csrf="+url.QueryEscape(session.CSRF()), nil)
	r = withCSRFTestSession(r, session)

	handler := newCSRFMiddleware("").handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal(`The next handler should not be called`)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf(`Unexpected status code, got: %d`, w.Code)
	}
}

func TestCSRFMiddlewareRemovesMultipartTemporaryFiles(t *testing.T) {
	// os.TempDir reads TMPDIR on Unix and TMP on Windows.
	tempDir := t.TempDir()
	t.Setenv("TMPDIR", tempDir)
	t.Setenv("TMP", tempDir)

	session := newCSRFTestSession(true)
	body, contentType := newMultipartBody(t, nil, bytes.Repeat([]byte("a"), 4096))

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", contentType)
	r.Header.Set("X-Csrf-Token", session.CSRF())
	r = withCSRFTestSession(r, session)

	handler := newCSRFMiddleware("").handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A tiny memory limit forces the file part to be written to disk.
		if err := r.ParseMultipartForm(1); err != nil {
			t.Fatalf(`Unable to parse the multipart form: %v`, err)
		}

		entries, err := os.ReadDir(tempDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 0 {
			t.Fatal(`Expected the file part to be stored in a temporary file`)
		}
	}))

	handler.ServeHTTP(httptest.NewRecorder(), r)

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf(`Expected temporary files to be removed, found %d`, len(entries))
	}
}
