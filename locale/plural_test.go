// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

import "testing"

func TestPluralRules(t *testing.T) {
	scenarios := map[string]map[int]int{
		"default": {
			1: 0,
			2: 1,
			5: 1,
		},
		"ar_AR": {
			0:   0,
			1:   1,
			2:   2,
			5:   3,
			11:  4,
			200: 5,
		},
		"cs_CZ": {
			1: 0,
			2: 1,
			5: 2,
		},
		"pl_PL": {
			1: 0,
			2: 1,
			5: 2,
		},
		"pt_BR": {
			1: 0,
			2: 1,
			5: 1,
		},
		"ru_RU": {
			1: 0,
			2: 1,
			5: 2,
		},
		"sr_RS": {
			1: 0,
			2: 1,
			5: 2,
		},
		"zh_CN": {
			1: 0,
			5: 0,
		},
	}

	for rule, values := range scenarios {
		for input, expected := range values {
			result := pluralForms[rule](input)
			if result != expected {
				t.Errorf(`Unexpected result for %q rule, got %d instead of %d for %d as input`, rule, result, expected, input)
			}
		}
	}
}
