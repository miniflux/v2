// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/v2/internal/locale"

// See https://localization-guide.readthedocs.io/en/latest/l10n/pluralforms.html
// And http://www.unicode.org/cldr/charts/29/supplemental/language_plural_rules.html
func getPluralForm(lang string, n int) int {
	switch lang {
	case "ar_AR":
		switch {
		case n == 0:
			return 0
		case n == 1:
			return 1
		case n == 2:
			return 2
		case n%100 >= 3 && n%100 <= 10:
			return 3
		case n%100 >= 11:
			return 4
		default:
			return 5
		}
	case "cs_CZ":
		switch {
		case n == 1:
			return 0
		case n >= 2 && n <= 4:
			return 1
		default:
			return 2
		}
	case "id_ID", "ja_JP":
		return 0
	case "pl_PL":
		switch {
		case n == 1:
			return 0
		case n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20):
			return 1
		default:
			return 2
		}
	case "ro_RO":
		switch {
		case n == 1:
			return 0
		case n == 0 || (n%100 > 0 && n%100 < 20):
			return 1
		default:
			return 2
		}
	case "ru_RU", "uk_UA", "sr_RS":
		switch {
		case n%10 == 1 && n%100 != 11:
			return 0
		case n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20):
			return 1
		default:
			return 2
		}
	case "zh_CN", "zh_TW", "nan_Latn_pehoeji":
		return 0
	default: // includes fr_FR, pr_BR, tr_TR
		if n > 1 {
			return 1
		}
		return 0
	}
}
