// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"bytes"
	"mime"
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
		NewBuilder(w, r).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	headers := map[string]string{
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
		NewBuilder(w, r).WithStatus(http.StatusNotAcceptable).Write()
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
		NewBuilder(w, r).WithHeader("X-My-Header", "Value").Write()
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
		NewBuilder(w, r).WithAttachment("my_file.pdf").Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := "attachment; filename=my_file.pdf"
	actual := resp.Header.Get("Content-Disposition")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}
}

func TestBuildResponseWithAttachmentEscapesFilename(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewBuilder(w, r).WithAttachment(`a";filename="malware.exe`).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	actual := resp.Header.Get("Content-Disposition")
	mediaType, params, err := mime.ParseMediaType(actual)
	if err != nil {
		t.Fatalf("Unexpected parse error for %q: %v", actual, err)
	}

	if mediaType != "attachment" {
		t.Fatalf(`Unexpected media type, got %q instead of %q`, mediaType, "attachment")
	}

	if params["filename"] != `a";filename="malware.exe` {
		t.Fatalf(`Unexpected filename, got %q instead of %q`, params["filename"], `a";filename="malware.exe`)
	}
}

func TestBuildResponseWithInlineEscapesFilename(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewBuilder(w, r).WithInline(`a";filename="malware.exe`).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	actual := resp.Header.Get("Content-Disposition")
	mediaType, params, err := mime.ParseMediaType(actual)
	if err != nil {
		t.Fatalf("Unexpected parse error for %q: %v", actual, err)
	}

	if mediaType != "inline" {
		t.Fatalf(`Unexpected media type, got %q instead of %q`, mediaType, "inline")
	}

	if params["filename"] != `a";filename="malware.exe` {
		t.Fatalf(`Unexpected filename, got %q instead of %q`, params["filename"], `a";filename="malware.exe`)
	}
}

func TestFormatContentDisposition(t *testing.T) {
	tests := []struct {
		name            string
		dispositionType string
		filename        string
		expected        string
	}{
		{"empty filename returns bare type", "inline", "", "inline"},
		{"simple filename", "attachment", "photo.jpg", `attachment; filename=photo.jpg`},
		{"filename with double quote", "inline", `a";filename="malware.exe`, `inline; filename="a\";filename=\"malware.exe"`},
		{"filename with spaces", "attachment", "my file.txt", `attachment; filename="my file.txt"`},
		{"non-ASCII filename", "attachment", "café.png", `attachment; filename*=utf-8''caf%C3%A9.png`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := formatContentDisposition(tt.dispositionType, tt.filename)
			if actual != tt.expected {
				t.Fatalf(`formatContentDisposition(%q, %q) = %q, want %q`, tt.dispositionType, tt.filename, actual, tt.expected)
			}
		})
	}
}

func TestBuildResponseWithByteBody(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewBuilder(w, r).WithBodyAsBytes([]byte("body")).Write()
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
		NewBuilder(w, r).WithCaching("etag", 1*time.Minute, func(b *Builder) {
			b.WithBodyAsString("cached body")
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

	expectedHeader := "public, immutable"
	actualHeader := resp.Header.Get("Cache-Control")
	if actualHeader != expectedHeader {
		t.Fatalf(`Unexpected cache control header, got %q instead of %q`, actualHeader, expectedHeader)
	}

	if actualETag := resp.Header.Get("ETag"); actualETag != `"etag"` {
		t.Fatalf(`Unexpected etag header, got %q instead of %q`, actualETag, `"etag"`)
	}

	if resp.Header.Get("Expires") == "" {
		t.Fatalf(`Expires header should not be empty`)
	}
}

func TestBuildResponseWithCachingAndIfNoneMatch(t *testing.T) {
	tests := []struct {
		name           string
		ifNoneMatch    string
		expectedStatus int
		expectedBody   string
	}{
		{"matching strong etag", `"etag"`, http.StatusNotModified, ""},
		{"matching weak etag", `W/"etag"`, http.StatusNotModified, ""},
		{"multiple etags with match", `"other", W/"etag"`, http.StatusNotModified, ""},
		{"wildcard", `*`, http.StatusNotModified, ""},
		{"non-matching etag", `"different"`, http.StatusOK, "cached body"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("If-None-Match", tt.ifNoneMatch)

			w := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				NewBuilder(w, r).WithCaching("etag", 1*time.Minute, func(b *Builder) {
					b.WithBodyAsString("cached body")
					b.Write()
				})
			})

			handler.ServeHTTP(w, r)
			resp := w.Result()

			if resp.StatusCode != tt.expectedStatus {
				t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, tt.expectedStatus)
			}

			if actual := w.Body.String(); actual != tt.expectedBody {
				t.Fatalf(`Unexpected body, got %q instead of %q`, actual, tt.expectedBody)
			}

			if resp.Header.Get("Cache-Control") != "public, immutable" {
				t.Fatalf(`Unexpected Cache-Control header: %q`, resp.Header.Get("Cache-Control"))
			}

			if resp.Header.Get("Expires") == "" {
				t.Fatalf(`Expires header should not be empty`)
			}
		})
	}
}

func TestNormalizeETag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc", `"abc"`},
		{`"already-quoted"`, `"already-quoted"`},
		{`W/"weak"`, `W/"weak"`},
		{"", ""},
		{"  spaced  ", `"spaced"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if actual := normalizeETag(tt.input); actual != tt.expected {
				t.Fatalf(`normalizeETag(%q) = %q, want %q`, tt.input, actual, tt.expected)
			}
		})
	}
}

func TestIfNoneMatch(t *testing.T) {
	tests := []struct {
		name        string
		headerValue string
		etag        string
		expected    bool
	}{
		{"empty header", "", `"etag"`, false},
		{"empty etag", `"etag"`, "", false},
		{"exact match", `"etag"`, `"etag"`, true},
		{"weak vs strong match", `W/"etag"`, `"etag"`, true},
		{"wildcard", `*`, `"etag"`, true},
		{"no match", `"other"`, `"etag"`, false},
		{"match in list", `"a", "etag", "b"`, `"etag"`, true},
		{"no match in list", `"a", "b", "c"`, `"etag"`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if actual := ifNoneMatch(tt.headerValue, tt.etag); actual != tt.expected {
				t.Fatalf(`ifNoneMatch(%q, %q) = %v, want %v`, tt.headerValue, tt.etag, actual, tt.expected)
			}
		})
	}
}

func TestBuildResponseWithBrotliCompression(t *testing.T) {
	body := strings.Repeat("a", compressionThreshold+1)
	r, err := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip, deflate, br")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewBuilder(w, r).WithBodyAsString(body).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := "br"
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}
}

func TestBuildResponseWithGzipCompression(t *testing.T) {
	body := strings.Repeat("a", compressionThreshold+1)
	r, err := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip, deflate")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewBuilder(w, r).WithBodyAsString(body).Write()
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
		NewBuilder(w, r).WithBodyAsString(body).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := "deflate"
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}

	expectedVary := "Accept-Encoding"
	actualVary := resp.Header.Get("Vary")
	if actualVary != expectedVary {
		t.Fatalf(`Unexpected vary header value, got %q instead of %q`, actualVary, expectedVary)
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
		NewBuilder(w, r).WithBodyAsString(body).WithoutCompression().Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := ""
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}

	expectedVary := ""
	actualVary := resp.Header.Get("Vary")
	if actualVary != expectedVary {
		t.Fatalf(`Unexpected vary header value, got %q instead of %q`, actualVary, expectedVary)
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
		NewBuilder(w, r).WithBodyAsString(body).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := ""
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}

	expectedVary := ""
	actualVary := resp.Header.Get("Vary")
	if actualVary != expectedVary {
		t.Fatalf(`Unexpected vary header value, got %q instead of %q`, actualVary, expectedVary)
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
		NewBuilder(w, r).WithBodyAsString(body).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expected := ""
	actual := resp.Header.Get("Content-Encoding")
	if actual != expected {
		t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
	}

	expectedVary := "Accept-Encoding"
	actualVary := resp.Header.Get("Vary")
	if actualVary != expectedVary {
		t.Fatalf(`Unexpected vary header value, got %q instead of %q`, actualVary, expectedVary)
	}
}

func TestBuildResponseWithReaderBody(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewBuilder(w, r).WithBodyAsReader(bytes.NewBufferString("body")).Write()
	})

	handler.ServeHTTP(w, r)

	if actualBody := w.Body.String(); actualBody != "body" {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, "body")
	}
}
