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
