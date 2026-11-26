// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pinboard

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCreateBookmarkErrorsDoNotContainAuthToken(t *testing.T) {
	const authToken = "user:0123456789ABCDEF"

	closeConnection := func(t *testing.T, w http.ResponseWriter) {
		connection, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Fatalf("unable to hijack connection: %v", err)
		}
		connection.Close()
	}

	notFound := func(w http.ResponseWriter) {
		w.Write([]byte(`<posts></posts>`))
	}

	tests := []struct {
		name            string
		handler         func(t *testing.T, w http.ResponseWriter, r *http.Request)
		wantErrContains string
	}{
		{
			name: "lookup connection fails",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				closeConnection(t, w)
			},
			wantErrContains: "pinboard: unable fetch bookmark",
		},
		{
			name: "lookup returns an error status",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantErrContains: "status=401",
		},
		{
			name: "add connection fails",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/posts/get" {
					notFound(w)
					return
				}
				closeConnection(t, w)
			},
			wantErrContains: "pinboard: unable to send request",
		},
		{
			name: "add returns an error status",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/posts/get" {
					notFound(w)
					return
				}
				w.WriteHeader(http.StatusTooManyRequests)
			},
			wantErrContains: "status=429",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("auth_token"); got != authToken {
					t.Errorf("auth_token = %q, want %q", got, authToken)
				}
				tt.handler(t, w, r)
			}))
			defer server.Close()

			client := NewClient(authToken)
			client.apiBaseURL = server.URL

			err := client.CreateBookmark("https://example.org/article", "Article", "tag", false)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErrContains)
			}
			for _, secret := range []string{authToken, url.QueryEscape(authToken)} {
				if strings.Contains(err.Error(), secret) {
					t.Errorf("error contains the auth token: %q", err)
				}
			}
		})
	}
}
