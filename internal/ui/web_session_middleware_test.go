// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"miniflux.app/v2/internal/http/request"
)

func TestWebSessionMiddlewareRedirectsWithoutCreatingSession(t *testing.T) {
	scenarios := []struct {
		name   string
		cookie *http.Cookie
	}{
		{"no cookie", nil},
		{"malformed cookie", &http.Cookie{Name: sessionCookieName, Value: "invalid"}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// A nil store makes the test fail if the middleware tries to persist a session.
			middleware := newWebSessionMiddleware("/base", nil)
			handler := middleware.handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal(`The next handler should not be called`)
			}))

			r := httptest.NewRequest(http.MethodGet, "/unread?page=2", nil)
			if scenario.cookie != nil {
				r.AddCookie(scenario.cookie)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			if w.Code != http.StatusFound {
				t.Fatalf(`Unexpected status code, got: %d`, w.Code)
			}

			if location := w.Header().Get("Location"); location != "/base/?redirect_url=%2Funread%3Fpage%3D2" {
				t.Fatalf(`Unexpected redirect location, got: %q`, location)
			}

			if cookie := w.Header().Get("Set-Cookie"); cookie != "" {
				t.Fatalf(`No session cookie should be set, got: %q`, cookie)
			}
		})
	}
}

func TestWebSessionMiddlewareSkipsStaticAssets(t *testing.T) {
	middleware := newWebSessionMiddleware("", nil)

	called := false
	handler := middleware.handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/favicon.ico", nil))

	if !called {
		t.Fatal(`The next handler should be called for static assets`)
	}

	if cookie := w.Header().Get("Set-Cookie"); cookie != "" {
		t.Fatalf(`No session cookie should be set, got: %q`, cookie)
	}
}

func TestWebSessionMiddlewareUsesInMemorySessionForPublicRoutes(t *testing.T) {
	scenarios := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/share/code"},
		{http.MethodGet, "/manifest.json"},
		{http.MethodGet, "/offline"},
		{http.MethodGet, "/oauth2/google/callback"},
		{http.MethodPost, "/webauthn/login/finish"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.method+" "+scenario.path, func(t *testing.T) {
			configureTestOptions(t, map[string]string{"WEBAUTHN": "1", "OAUTH2_PROVIDER": "google"})

			// A nil store makes the test fail if the middleware tries to persist a session.
			middleware := newWebSessionMiddleware("", nil)

			called := false
			handler := middleware.handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true

				session := request.WebSession(r)
				if session == nil {
					t.Fatal(`A web session should be available in the request context`)
				}

				if session.Language() != "fr_FR" || session.Theme() != "dark_serif" {
					t.Errorf(`Preferences should be seeded from the cookie, got %q and %q`, session.Language(), session.Theme())
				}

				// Changes to an in-memory session must not be persisted either.
				session.SetErrorMessage("error")
			}))

			r := httptest.NewRequest(scenario.method, scenario.path, nil)
			r.AddCookie(&http.Cookie{Name: preferencesCookieName, Value: "fr_FR.dark_serif"})
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			if !called {
				t.Fatal(`The next handler should be called for public routes`)
			}

			if cookie := w.Header().Get("Set-Cookie"); cookie != "" {
				t.Fatalf(`No session cookie should be set, got: %q`, cookie)
			}
		})
	}
}
