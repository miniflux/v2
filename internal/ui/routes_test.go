// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsStaticAssetRoute(t *testing.T) {
	scenarios := []struct {
		path     string
		expected bool
	}{
		{"/favicon.ico", true},
		{"/robots.txt", true},
		{"/stylesheets/checksum/light_serif.css", true},
		{"/js/checksum/app.js", true},
		{"/icon/checksum/icon-120.png", true},
		{"/feed-icon/external-id", true},
		{"/proxy/digest/url", true},
		{"/", false},
		{"/login", false},
		{"/manifest.json", false},
		{"/share/code", false},
		{"/unread", false},
	}

	for _, scenario := range scenarios {
		r := httptest.NewRequest(http.MethodGet, scenario.path, nil)
		if result := isStaticAssetRoute(r); result != scenario.expected {
			t.Errorf(`Unexpected result for %q, got %v instead of %v`, scenario.path, result, scenario.expected)
		}
	}
}

func TestIsSessionStartRoute(t *testing.T) {
	scenarios := []struct {
		name     string
		env      map[string]string
		method   string
		path     string
		expected bool
	}{
		{"login page", nil, http.MethodGet, "/", true},
		{"login page head", nil, http.MethodHead, "/", true},
		{"login page post", nil, http.MethodPost, "/", false},
		{"login form submission", nil, http.MethodPost, "/login", false},
		{"shared entry", nil, http.MethodGet, "/share/code", false},
		{"web manifest", nil, http.MethodGet, "/manifest.json", false},
		{"offline page", nil, http.MethodGet, "/offline", false},
		{"webauthn begin disabled", nil, http.MethodGet, "/webauthn/login/begin", false},
		{"webauthn begin enabled", map[string]string{"WEBAUTHN": "1"}, http.MethodGet, "/webauthn/login/begin", true},
		{"webauthn finish", map[string]string{"WEBAUTHN": "1"}, http.MethodPost, "/webauthn/login/finish", false},
		{"oauth2 redirect disabled", nil, http.MethodGet, "/oauth2/google/redirect", false},
		{"oauth2 redirect enabled", map[string]string{"OAUTH2_PROVIDER": "google"}, http.MethodGet, "/oauth2/google/redirect", true},
		{"oauth2 callback", map[string]string{"OAUTH2_PROVIDER": "google"}, http.MethodGet, "/oauth2/google/callback", false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			configureTestOptions(t, scenario.env)

			r := httptest.NewRequest(scenario.method, scenario.path, nil)
			if result := isSessionStartRoute(r); result != scenario.expected {
				t.Errorf(`Unexpected result for %s %q, got %v instead of %v`, scenario.method, scenario.path, result, scenario.expected)
			}
		})
	}
}
