// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever // import "miniflux.app/v2/internal/fever"

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareRejectsMissingAPIKey(t *testing.T) {
	newPostRequest := func(body string) *http.Request {
		request := httptest.NewRequest(http.MethodPost, "/fever/?api", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return request
	}

	scenarios := []struct {
		name    string
		request *http.Request
	}{
		{"GET without api_key", httptest.NewRequest(http.MethodGet, "/fever/?api&groups", nil)},
		{"GET with an empty api_key", httptest.NewRequest(http.MethodGet, "/fever/?api&groups&api_key=", nil)},
		{"POST without api_key", newPostRequest("mark=item&as=read&id=1")},
		{"POST with an empty api_key", newPostRequest("api_key=&mark=item&as=read&id=1")},
	}

	// The store is nil: every request must be rejected before the store is queried.
	middleware := Middleware(nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			middleware(next).ServeHTTP(recorder, scenario.request)

			if recorder.Code != http.StatusOK {
				t.Fatalf(`Expected status %d, got %d`, http.StatusOK, recorder.Code)
			}

			var result baseResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
				t.Fatalf(`Unable to decode %s: %v`, recorder.Body.Bytes(), err)
			}
			if result.Version != 3 || result.Authenticated != 0 {
				t.Errorf(`Expected an authentication failure for API version 3, got %s`, recorder.Body.Bytes())
			}
		})
	}
}
