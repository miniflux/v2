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

func TestAuthMiddlewareRejectsMalformedCredentials(t *testing.T) {
	newGetRequest := func(target, authorization string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, target, nil)
		if authorization != "" {
			r.Header.Set("Authorization", authorization)
		}
		return r
	}

	newPostRequest := func(body, authorization string) *http.Request {
		r := httptest.NewRequest(http.MethodPost, "/reader/api/0/edit-tag", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if authorization != "" {
			r.Header.Set("Authorization", authorization)
		}
		return r
	}

	scenarios := []struct {
		name    string
		request *http.Request
	}{
		{"GET without Authorization header", newGetRequest("/reader/api/0/user-info", "")},
		{"GET with the token in the query string", newGetRequest("/reader/api/0/user-info?T=reader/abc", "")},
		{"GET with a scheme only", newGetRequest("/reader/api/0/user-info", "GoogleLogin")},
		{"GET with another scheme", newGetRequest("/reader/api/0/user-info", "Bearer auth=reader/abc")},
		{"GET with a capitalized auth field", newGetRequest("/reader/api/0/user-info", "GoogleLogin Auth=reader/abc")},
		{"GET without an auth value", newGetRequest("/reader/api/0/user-info", "GoogleLogin auth")},
		{"GET with several equal signs", newGetRequest("/reader/api/0/user-info", "GoogleLogin auth=reader/abc=")},
		{"GET with an extra field", newGetRequest("/reader/api/0/user-info", "GoogleLogin auth=reader/abc extra")},
		{"GET with a token without a slash", newGetRequest("/reader/api/0/user-info", "GoogleLogin auth=readerabc")},
		{"GET with a token with several slashes", newGetRequest("/reader/api/0/user-info", "GoogleLogin auth=reader/abc/def")},
		{"GET with a token without a username", newGetRequest("/reader/api/0/user-info", "GoogleLogin auth=/abc")},
		{"POST without a T field", newPostRequest(url.Values{"i": {"1"}}.Encode(), "")},
		{"POST with the token in the Authorization header", newPostRequest("", "GoogleLogin auth=reader/abc")},
		{"POST with an empty T field", newPostRequest("T=", "")},
		{"POST with a token without a slash", newPostRequest("T=readerabc", "")},
		{"POST with a token without a username", newPostRequest("T=%2Fabc", "")},
		{"POST with an invalid form body", newPostRequest("T=%zz", "")},
	}

	// The store is nil: every request must be rejected before the store is queried.
	middleware := newAuthMiddleware(nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			middleware.validateApiKey(next).ServeHTTP(recorder, scenario.request)
			checkUnauthorizedResponse(t, recorder)
		})
	}
}
