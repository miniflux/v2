// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package view

import (
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestCSPExternalFont(t *testing.T) {
	want := []string{
		`default-src 'none';`,
		`img-src * data:;`,
		`media-src *;`,
		`frame-src *;`,
		`style-src 'nonce-1234';`,
		`script-src 'nonce-1234'`,
		`'strict-dynamic';`,
		`font-src test.com;`,
		`require-trusted-types-for 'script';`,
		`trusted-types html url;`,
		`manifest-src 'self';`,
	}
	got := csp(&model.User{ExternalFontHosts: "test.com"}, "1234")

	for _, value := range want {
		if !strings.Contains(got, value) {
			t.Errorf(`Unexpected result, didn't find %q in %q`, value, got)
		}
	}
}

func TestCSPNoUser(t *testing.T) {
	want := []string{
		`default-src 'none';`,
		`img-src * data:;`,
		`media-src *;`,
		`frame-src *;`,
		`style-src 'nonce-1234';`,
		`script-src 'nonce-1234'`,
		`'strict-dynamic';`,
		`require-trusted-types-for 'script';`,
		`trusted-types html url;`,
		`manifest-src 'self';`,
	}
	got := csp(nil, "1234")

	for _, value := range want {
		if !strings.Contains(got, value) {
			t.Errorf(`Unexpected result, didn't find %q in %q`, value, got)
		}
	}
}

func TestCSPCustomJSExternalFont(t *testing.T) {
	want := []string{
		`default-src 'none';`,
		`img-src * data:;`,
		`media-src *;`,
		`frame-src *;`,
		`style-src 'nonce-1234';`,
		`script-src 'nonce-1234'`,
		`'strict-dynamic';`,
		`require-trusted-types-for 'script';`,
		`trusted-types html url;`,
		`manifest-src 'self';`,
	}
	got := csp(&model.User{ExternalFontHosts: "test.com", CustomJS: "alert(1)"}, "1234")

	for _, value := range want {
		if !strings.Contains(got, value) {
			t.Errorf(`Unexpected result, didn't find %q in %q`, value, got)
		}
	}
}

func TestCSPExternalFontStylesheet(t *testing.T) {
	want := []string{
		`default-src 'none';`,
		`img-src * data:;`,
		`media-src *;`,
		`frame-src *;`,
		`style-src 'nonce-1234' test.com;`,
		`script-src 'nonce-1234'`,
		`'strict-dynamic';`,
		`require-trusted-types-for 'script';`,
		`trusted-types html url;`,
		`manifest-src 'self';`,
	}
	got := csp(&model.User{ExternalFontHosts: "test.com", Stylesheet: "a {color: red;}"}, "1234")

	for _, value := range want {
		if !strings.Contains(got, value) {
			t.Errorf(`Unexpected result, didn't find %q in %q`, value, got)
		}
	}
}

// CSP must not come back as a meta tag: it ships as a response header.
// See deadhead rule meta/http-equiv-content-security-policy (harmful).
func TestCSPDoesNotReturnMetaTag(t *testing.T) {
	got := csp(nil, "1234")

	lower := strings.ToLower(got)
	if strings.Contains(lower, "<meta") || strings.Contains(lower, "http-equiv") {
		t.Fatalf("csp() must return a header value, not a meta tag, got %q", got)
	}
}
