// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer // import "miniflux.app/v2/internal/reader/sanitizer"

import "testing"

func TestHasValidURIScheme(t *testing.T) {
	scenarios := map[string]bool{
		// Allowed: web schemes.
		"http://example.org/article":  true,
		"https://example.org/article": true,

		// Allowed: a sample of the broader feed-content schemes.
		"mailto:author@example.org": true,
		"magnet:?xt=urn:btih:abc":   true,
		"tel:+15551234567":          true,
		"ftp://example.org/file":    true,
		"feed:https://example.org/": true,
		"webcal://example.org/cal":  true,

		// Rejected: schemes that enable script execution or local resource access.
		"javascript:alert(1)":                      false,
		"data:text/html,<script>alert(1)</script>": false,
		"vbscript:msgbox(1)":                       false,
		"file:///etc/passwd":                       false,

		// Rejected: missing or malformed scheme.
		"":                        false,
		"example.org":             false,
		"/relative/path":          false,
		"//evil.example.org/path": false,

		// Rejected: case-sensitive match (callers are expected to pass
		// already-normalized URLs, e.g. via net/url which lowercases the scheme).
		"HTTPS://example.org": false,
		"JavaScript:alert(1)": false,
	}

	for input, expected := range scenarios {
		t.Run(input, func(t *testing.T) {
			if actual := HasValidURIScheme(input); actual != expected {
				t.Errorf("HasValidURIScheme(%q) = %v, want %v", input, actual, expected)
			}
		})
	}
}
