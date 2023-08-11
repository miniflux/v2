// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResponseHasCommonHeaders(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	headers := map[string]string{
		"X-XSS-Protection":       "1; mode=block",
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
	}

	for header, expected := range headers {
		actual := resp.Header.Get(header)
		if actual != expected {
			t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
		}
	}
}

func TestBuildResponseWithCustomStatusCode(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithStatus(http.StatusNotAcceptable).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusNotAcceptable
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}
}

func TestBuildResponseWithCustomHeader(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithHeader("X-My-Header", "Value").Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := "Value"
	actual := resp.Header.Get("X-My-Header")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}
}

func TestBuildResponseWithAttachment(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithAttachment("my_file.pdf").Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := "attachment; filename=my_file.pdf"
	actual := resp.Header.Get("Content-Disposition")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}
}

func TestBuildResponseWithError(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithBody(errors.New("Some error")).Write()
	})

	handler.ServeHTTP(w, r)

	expectedBody := `Some error`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}
}

func TestBuildResponseWithByteBody(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithBody([]byte("body")).Write()
	})

	handler.ServeHTTP(w, r)

	expectedBody := `body`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}
}

func TestBuildResponseWithCachingEnabled(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithCaching("etag", 1*time.Minute, func(b *Builder) {
			b.WithBody("cached body")
			b.Write()
		})
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusOK
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `cached body`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedHeader := "public"
	actualHeader := resp.Header.Get("Cache-Control")
	if actualHeader != expectedHeader {
		t.Fatalf(`Unexpected cache control header, got %q instead of %q`, actualHeader, expectedHeader)
	}

	if resp.Header.Get("Expires") == "" {
		t.Fatalf(`Expires header should not be empty`)
	}
}

func TestBuildResponseWithCachingAndEtag(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	r.Header.Set("If-None-Match", "etag")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithCaching("etag", 1*time.Minute, func(b *Builder) {
			b.WithBody("cached body")
			b.Write()
		})
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusNotModified
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := ``
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedHeader := "public"
	actualHeader := resp.Header.Get("Cache-Control")
	if actualHeader != expectedHeader {
		t.Fatalf(`Unexpected cache control header, got %q instead of %q`, actualHeader, expectedHeader)
	}

	if resp.Header.Get("Expires") == "" {
		t.Fatalf(`Expires header should not be empty`)
	}
}

func TestBuildResponseWithGzipCompression(t *testing.T) {
	body := strings.Repeat("a", compressionThreshold+1)
	r, err := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip, deflate, br")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithBody(body).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := "gzip"
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}
}

func TestBuildResponseWithDeflateCompression(t *testing.T) {
	body := strings.Repeat("a", compressionThreshold+1)
	r, err := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "deflate")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithBody(body).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := "deflate"
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}
}

func TestBuildResponseWithCompressionDisabled(t *testing.T) {
	body := strings.Repeat("a", compressionThreshold+1)
	r, err := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "deflate")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithBody(body).WithoutCompression().Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := ""
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}
}

func TestBuildResponseWithDeflateCompressionAndSmallPayload(t *testing.T) {
	body := strings.Repeat("a", compressionThreshold)
	r, err := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "deflate")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithBody(body).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := ""
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}
}

func TestBuildResponseWithoutCompressionHeader(t *testing.T) {
	body := strings.Repeat("a", compressionThreshold+1)
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).WithBody(body).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := ""
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}
}
