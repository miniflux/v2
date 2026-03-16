// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestXMLResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		XML(w, r, "Some XML")
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusOK)
	}

	if actualBody := w.Body.String(); actualBody != "Some XML" {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, "Some XML")
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != "text/xml; charset=utf-8" {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, "text/xml; charset=utf-8")
	}
}

func TestXMLAttachmentResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		XMLAttachment(w, r, "file.xml", "Some XML")
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusOK)
	}

	if actualBody := w.Body.String(); actualBody != "Some XML" {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, "Some XML")
	}

	headers := map[string]string{
		"Content-Type":        "text/xml; charset=utf-8",
		"Content-Disposition": "attachment; filename=file.xml",
	}

	for header, expected := range headers {
		if actual := resp.Header.Get(header); actual != expected {
			t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
		}
	}
}
