// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package language // import "miniflux.app/v2/internal/reader/language"

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"   ", ""},
		{"en", "en"},
		{"EN", "en"},
		{"en_US", "en-us"},
		{"EN-us", "en-us"},
		{"pt-BR", "pt-br"},
		{"  fr-FR  ", "fr-fr"},
	}
	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
