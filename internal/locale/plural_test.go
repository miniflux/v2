// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/v2/internal/locale"

import "testing"

func TestPluralRules(t *testing.T) {
	scenarios := map[string]map[int]int{
		// Default rule (covers fr_FR, pt_BR, tr_TR, and other unlisted languages)
		"default": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
			5: 1, // n > 1
		},
		// Arabic (ar_AR) - 6 forms
		"ar_AR": {
			0:   0, // n == 0
			1:   1, // n == 1
			2:   2, // n == 2
			3:   3, // n%100 >= 3 && n%100 <= 10
			5:   3, // n%100 >= 3 && n%100 <= 10
			10:  3, // n%100 >= 3 && n%100 <= 10
			11:  4, // n%100 >= 11
			15:  4, // n%100 >= 11
			99:  4, // n%100 >= 11
			100: 5, // default case (n%100 == 0, doesn't match any condition)
			101: 5, // default case (n%100 == 1, but n != 1)
			200: 5, // default case
		},
		// Czech (cs_CZ) - 3 forms
		"cs_CZ": {
			1: 0, // n == 1
			2: 1, // n >= 2 && n <= 4
			3: 1, // n >= 2 && n <= 4
			4: 1, // n >= 2 && n <= 4
			5: 2, // default case
		},
		// French (fr_FR) - uses default rule
		"fr_FR": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
			5: 1, // n > 1
		},
		// Indonesian (id_ID) - always form 0
		"id_ID": {
			0:   0,
			1:   0,
			5:   0,
			100: 0,
		},
		// Japanese (ja_JP) - always form 0
		"ja_JP": {
			0:   0,
			1:   0,
			2:   0,
			5:   0,
			100: 0,
		},
		// Polish (pl_PL) - 3 forms
		"pl_PL": {
			1:  0, // n == 1
			2:  1, // n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20)
			3:  1, // n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20)
			4:  1, // n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20)
			5:  2, // default case
			10: 2, // default case (n%100 < 10, but n%10 not in 2-4)
			11: 2, // default case (n%100 >= 10 and < 20)
			12: 2, // default case (n%100 >= 10 and < 20)
			22: 1, // n%10 >= 2 && n%10 <= 4 && (n%100 >= 20)
			24: 1, // n%10 >= 2 && n%10 <= 4 && (n%100 >= 20)
		},
		// Portuguese Brazilian (pt_BR) - uses default rule
		"pt_BR": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
			5: 1, // n > 1
		},
		// Romanian (ro_RO) - 3 forms
		"ro_RO": {
			0:   1, // n == 0 || (n%100 > 0 && n%100 < 20)
			1:   0, // n == 1
			2:   1, // n == 0 || (n%100 > 0 && n%100 < 20)
			5:   1, // n == 0 || (n%100 > 0 && n%100 < 20)
			19:  1, // n == 0 || (n%100 > 0 && n%100 < 20)
			20:  2, // default case
			21:  2, // default case
			100: 2, // default case (n%100 == 0, so condition fails)
			101: 1, // n%100 == 1, so n%100 > 0 && n%100 < 20
		},
		// Russian (ru_RU) - 3 forms
		"ru_RU": {
			1:  0, // n%10 == 1 && n%100 != 11
			2:  1, // n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20)
			3:  1, // n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20)
			4:  1, // n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20)
			5:  2, // default case
			11: 2, // n%10 == 1 but n%100 == 11, so default case
			12: 2, // default case
			21: 0, // n%10 == 1 && n%100 != 11
			22: 1, // n%10 >= 2 && n%10 <= 4 && (n%100 >= 20)
		},
		// Serbian (sr_RS) - same as Russian
		"sr_RS": {
			1:  0, // n%10 == 1 && n%100 != 11
			2:  1, // n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20)
			5:  2, // default case
			11: 2, // n%10 == 1 but n%100 == 11, so default case
			21: 0, // n%10 == 1 && n%100 != 11
		},
		// Turkish (tr_TR) - uses default rule
		"tr_TR": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
			5: 1, // n > 1
		},
		// Ukrainian (uk_UA) - same as Russian
		"uk_UA": {
			1:  0, // n%10 == 1 && n%100 != 11
			2:  1, // n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20)
			5:  2, // default case
			11: 2, // n%10 == 1 but n%100 == 11, so default case
			21: 0, // n%10 == 1 && n%100 != 11
		},
		// Chinese Simplified (zh_CN) - always form 0
		"zh_CN": {
			0:   0,
			1:   0,
			5:   0,
			100: 0,
		},
		// Chinese Traditional (zh_TW) - always form 0
		"zh_TW": {
			0:   0,
			1:   0,
			5:   0,
			100: 0,
		},
		// Min Nan (nan_Latn_pehoeji) - always form 0
		"nan_Latn_pehoeji": {
			0:   0,
			1:   0,
			5:   0,
			100: 0,
		},
		// Additional languages from AvailableLanguages that use default rule
		"de_DE": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
		},
		"el_EL": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
		},
		"en_US": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
		},
		"es_ES": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
		},
		"fi_FI": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
		},
		"hi_IN": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
		},
		"it_IT": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
		},
		"nl_NL": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
		},
		// Test a language not in the switch (should use default rule)
		"unknown_language": {
			0: 0, // n <= 1
			1: 0, // n <= 1
			2: 1, // n > 1
		},
	}

	for rule, values := range scenarios {
		for input, expected := range values {
			result := getPluralForm(rule, input)
			if result != expected {
				t.Errorf(`Unexpected result for %q rule, got %d instead of %d for %d as input`, rule, result, expected, input)
			}
		}
	}
}
