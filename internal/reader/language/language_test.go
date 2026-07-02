// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package language // import "miniflux.app/v2/internal/reader/language"

import (
	"strings"
	"testing"
)

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
		{"zh-hant-cn-x-private1-private2", "zh-hant-cn-x-private1-private2"},

		// Values outside the tag alphabet are rejected, not stripped.
		{"en US", ""},
		{"en-US, de-DE", ""},
		{"en\x00us", ""},
		{"en\u202eus", ""},
		{"français", ""},
		{`"><script>`, ""},

		// Non-ASCII input must be rejected even when Unicode case
		// folding would map it to ASCII (U+212A Kelvin sign -> "k",
		// U+0130 dotted capital I -> "i").
		{"KO", ""},
		{"İ-en", ""},

		// Values longer than 50 characters are rejected.
		{strings.Repeat("a", 51), ""},
		{"en-" + strings.Repeat("a", 100), ""},
	}
	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
