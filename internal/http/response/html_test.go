// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTMLResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HTML(w, r, "Some HTML")
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusOK)
	}

	if actualBody := w.Body.String(); actualBody != `Some HTML` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `Some HTML`)
	}

	headers := map[string]string{
		"Content-Type":  "text/html; charset=utf-8",
		"Cache-Control": "no-cache, max-age=0, must-revalidate, no-store",
	}

	for header, expected := range headers {
		if actual := resp.Header.Get(header); actual != expected {
			t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
		}
	}
}

func TestHTMLServerErrorResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HTMLServerError(w, r, errors.New("Some error with injected HTML <script>alert('XSS')</script>"))
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusInternalServerError)
	}

	if actualBody := w.Body.String(); actualBody != `Some error with injected HTML &lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `Some error with injected HTML &lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != "text/plain; charset=utf-8" {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, "text/plain; charset=utf-8")
	}
}

func TestHTMLBadRequestResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HTMLBadRequest(w, r, errors.New("Some error with injected HTML <script>alert('XSS')</script>"))
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusBadRequest)
	}

	if actualBody := w.Body.String(); actualBody != `Some error with injected HTML &lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `Some error with injected HTML &lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != "text/plain; charset=utf-8" {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, "text/plain; charset=utf-8")
	}
}

func TestHTMLForbiddenResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HTMLForbidden(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusForbidden)
	}

	if actualBody := w.Body.String(); actualBody != `Access Forbidden` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `Access Forbidden`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != "text/html; charset=utf-8" {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, "text/html; charset=utf-8")
	}
}

func TestHTMLNotFoundResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HTMLNotFound(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusNotFound)
	}

	if actualBody := w.Body.String(); actualBody != `Page Not Found` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `Page Not Found`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != "text/html; charset=utf-8" {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, "text/html; charset=utf-8")
	}
}

func TestHTMLRedirectResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HTMLRedirect(w, r, "/path")
	})

	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusFound)
	}

	if actualResult := resp.Header.Get("Location"); actualResult != "/path" {
		t.Fatalf(`Unexpected redirect location, got %q instead of %q`, actualResult, "/path")
	}
}

func TestHTMLRedirectAcceptedTargets(t *testing.T) {
	scenarios := []string{
		"/feeds",
		"/category/1/entries",
		"https://example.org/article",
		"http://example.org/article",
	}

	for _, target := range scenarios {
		t.Run(target, func(t *testing.T) {
			r, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			HTMLRedirect(w, r, target)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusFound {
				t.Fatalf(`Unexpected status code for %q, got %d instead of %d`, target, resp.StatusCode, http.StatusFound)
			}

			if actualResult := resp.Header.Get("Location"); actualResult != target {
				t.Fatalf(`Unexpected redirect location, got %q instead of %q`, actualResult, target)
			}
		})
	}
}

func TestHTMLRedirectRejectsUnsafeTargets(t *testing.T) {
	scenarios := []string{
		"javascript:alert(1)",
		"JAVASCRIPT:alert(1)",
		"data:text/html,<script>alert(1)</script>",
		"vbscript:msgbox(1)",
		"file:///etc/passwd",
		"mailto:victim@example.org",
		"//evil.example.org/path",
		"ftp://example.org/file",
		"",
	}

	for _, target := range scenarios {
		t.Run(target, func(t *testing.T) {
			r, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			HTMLRedirect(w, r, target)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf(`Expected 400 for %q, got %d`, target, resp.StatusCode)
			}

			if location := resp.Header.Get("Location"); location != "" {
				t.Fatalf(`Expected no Location header for %q, got %q`, target, location)
			}
		})
	}
}

func TestHTMLRequestedRangeNotSatisfiable(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HTMLRequestedRangeNotSatisfiable(w, r, "bytes */12777")
	})

	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusRequestedRangeNotSatisfiable)
	}

	if actualContentRangeHeader := resp.Header.Get("Content-Range"); actualContentRangeHeader != "bytes */12777" {
		t.Fatalf(`Unexpected content range header, got %q instead of %q`, actualContentRangeHeader, "bytes */12777")
	}
}
