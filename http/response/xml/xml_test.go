// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package xml // import "miniflux.app/http/response/xml"

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOKResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		OK(w, r, "Some XML")
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusOK
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `Some XML`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := "text/xml; charset=utf-8"
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestAttachmentResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Attachment(w, r, "file.xml", "Some XML")
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusOK
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `Some XML`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	headers := map[string]string{
		"Content-Type":        "text/xml; charset=utf-8",
		"Content-Disposition": "attachment; filename=file.xml",
	}

	for header, expected := range headers {
		actual := resp.Header.Get(header)
		if actual != expected {
			t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
		}
	}
}
