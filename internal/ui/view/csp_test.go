// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package view

import (
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestCSP(t *testing.T) {
	const (
		defaultPolicy = `default-src 'none'; connect-src 'self'; frame-src *; img-src * data:; manifest-src 'self'; media-src *; ` +
			`require-trusted-types-for 'script'; script-src 'nonce-1234' 'strict-dynamic'; style-src 'nonce-1234'; trusted-types html url`

		externalFontPolicy = `default-src 'none'; connect-src 'self'; font-src test.com; frame-src *; img-src * data:; manifest-src 'self'; media-src *; ` +
			`require-trusted-types-for 'script'; script-src 'nonce-1234' 'strict-dynamic'; style-src 'nonce-1234'; trusted-types html url`
	)

	scenarios := []struct {
		name string
		user *model.User
		want string
	}{
		{"no user", nil, defaultPolicy},
		{"no external font hosts", &model.User{}, defaultPolicy},
		{"stylesheet without external font hosts", &model.User{Stylesheet: "a {color: red;}"}, defaultPolicy},
		{"external font hosts", &model.User{ExternalFontHosts: "test.com"}, externalFontPolicy},
		{"custom JavaScript keeps the policy", &model.User{ExternalFontHosts: "test.com", CustomJS: "alert(1)"}, externalFontPolicy},
		{
			"external font hosts with a stylesheet",
			&model.User{ExternalFontHosts: "test.com", Stylesheet: "a {color: red;}"},
			`default-src 'none'; connect-src 'self'; font-src test.com; frame-src *; img-src * data:; manifest-src 'self'; media-src *; ` +
				`require-trusted-types-for 'script'; script-src 'nonce-1234' 'strict-dynamic'; style-src 'nonce-1234' test.com; trusted-types html url`,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if got := csp(scenario.user, "1234"); got != scenario.want {
				t.Errorf("Unexpected policy:\ngot  %q\nwant %q", got, scenario.want)
			}
		})
	}
}
