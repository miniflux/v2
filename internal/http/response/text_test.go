// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTextResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Text(w, r, "Some plain text")
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusOK)
	}

	if actualBody := w.Body.String(); actualBody != "Some plain text" {
		t.Fatalf(`Unexpected body, got %q instead of %q`, actualBody, "Some plain text")
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != "text/plain; charset=utf-8" {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, "text/plain; charset=utf-8")
	}
}
