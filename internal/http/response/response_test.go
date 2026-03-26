// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNoContentResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NoContent(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusNoContent)
	}

	if actualBody := w.Body.String(); actualBody != `` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, ``)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != "" {
		t.Fatalf(`Unexpected content type, got %q instead of empty string`, actualContentType)
	}
}
