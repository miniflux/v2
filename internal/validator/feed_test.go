// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import "testing"

func TestIsValidRefreshInterval(t *testing.T) {
	cases := []struct {
		name  string
		value int
		want  bool
	}{
		{"below minimum", MinFeedRefreshIntervalMinutes - 1, false},
		{"at minimum", MinFeedRefreshIntervalMinutes, true},
		{"normal", 60, true},
		{"at maximum", MaxFeedRefreshIntervalMinutes, true},
		{"above maximum", MaxFeedRefreshIntervalMinutes + 1, false},
		{"zero", 0, false},
		{"negative", -1, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isValidRefreshInterval(c.value); got != c.want {
				t.Errorf(`isValidRefreshInterval(%d) = %v, want %v`, c.value, got, c.want)
			}
		})
	}
}
