// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"miniflux.app/v2/internal/config"
)

func configureTestOptions(t *testing.T, env map[string]string) {
	t.Helper()

	for key, value := range env {
		t.Setenv(key, value)
	}

	parsedOptions, err := config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Unable to configure test options: %v`, err)
	}

	previousOptions := config.Opts
	config.Opts = parsedOptions
	t.Cleanup(func() {
		config.Opts = previousOptions
	})
}

func TestSetPreferencesCookie(t *testing.T) {
	scenarios := []struct {
		name       string
		env        map[string]string
		wantPath   string
		wantSecure bool
	}{
		{"defaults", map[string]string{}, "/", false},
		{"base path over HTTPS", map[string]string{"BASE_URL": "https://example.org/reader", "HTTPS": "1"}, "/reader", true},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			configureTestOptions(t, scenario.env)

			w := httptest.NewRecorder()
			setPreferencesCookie(w, "de_DE", "dark_serif")

			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf(`Expected one cookie, got %d`, len(cookies))
			}

			cookie := cookies[0]
			if cookie.Name != preferencesCookieName || cookie.Value != "de_DE.dark_serif" {
				t.Errorf(`Unexpected cookie: %s=%s`, cookie.Name, cookie.Value)
			}
			if cookie.Path != scenario.wantPath {
				t.Errorf(`Unexpected path, got %q, want %q`, cookie.Path, scenario.wantPath)
			}
			if cookie.Secure != scenario.wantSecure {
				t.Errorf(`Unexpected Secure attribute, got %v, want %v`, cookie.Secure, scenario.wantSecure)
			}
			if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
				t.Errorf(`Expected an HttpOnly cookie with SameSite=Lax`)
			}
		})
	}
}

func TestPreferencesFromCookie(t *testing.T) {
	scenarios := []struct {
		name         string
		value        string
		wantLanguage string
		wantTheme    string
		wantOK       bool
	}{
		{"valid", "de_DE.dark_serif", "de_DE", "dark_serif", true},
		{"unknown language", "xx_XX.dark_serif", "", "", false},
		{"unknown theme", "de_DE.neon", "", "", false},
		{"missing separator", "de_DE", "", "", false},
		{"empty", "", "", "", false},
		{"no cookie", "", "", "", false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if scenario.name != "no cookie" {
				r.AddCookie(&http.Cookie{Name: preferencesCookieName, Value: scenario.value})
			}

			language, theme, ok := preferencesFromCookie(r)
			if language != scenario.wantLanguage || theme != scenario.wantTheme || ok != scenario.wantOK {
				t.Errorf(`Got (%q, %q, %v), want (%q, %q, %v)`, language, theme, ok, scenario.wantLanguage, scenario.wantTheme, scenario.wantOK)
			}
		})
	}
}
