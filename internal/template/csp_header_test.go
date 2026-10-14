// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"strings"
	"testing"
)

// CSP must be delivered as an HTTP header, never as a meta tag.
// See deadhead rule meta/http-equiv-content-security-policy (harmful):
// meta policies drop frame-ancestors/report-uri/sandbox, stop the preload
// scanner, and apply late. The template helper must therefore return a
// header value, not a <meta> element.
func TestCSPDoesNotReturnMetaTag(t *testing.T) {
	got := CSPPolicy(nil, "1234")

	if strings.Contains(strings.ToLower(got), "<meta") {
		t.Fatalf("CSPPolicy() must not return a <meta> tag (use HTTP header instead), got %q", got)
	}

	if strings.Contains(strings.ToLower(got), "http-equiv") {
		t.Fatalf("CSPPolicy() must not return http-equiv content (use HTTP header instead), got %q", got)
	}

	for _, want := range []string{
		`default-src 'none';`,
		`script-src 'nonce-1234'`,
		`style-src 'nonce-1234';`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("CSPPolicy() = %q, want it to contain %q", got, want)
		}
	}
}
