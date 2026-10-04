// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthMiddlewareRejectsTokenWithEmptyUsername(t *testing.T) {
	token := getAuthToken("", "")

	getRequest := httptest.NewRequest(http.MethodGet, "/reader/api/0/user-info", nil)
	getRequest.Header.Set("Authorization", "GoogleLogin auth="+token)

	postRequest := httptest.NewRequest(http.MethodPost, "/reader/api/0/edit-tag", strings.NewReader(url.Values{"T": {token}}.Encode()))
	postRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// The store is nil: the middleware must reject the token before querying it.
	middleware := newAuthMiddleware(nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	for _, request := range []*http.Request{getRequest, postRequest} {
		recorder := httptest.NewRecorder()
		middleware.validateApiKey(next).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s: expected status %d, got %d", request.Method, http.StatusUnauthorized, recorder.Code)
		}
	}
}
