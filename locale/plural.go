// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

type pluralFormFunc func(n int) int

// See https://localization-guide.readthedocs.io/en/latest/l10n/pluralforms.html
// And http://www.unicode.org/cldr/charts/29/supplemental/language_plural_rules.html
var pluralForms = map[string]pluralFormFunc{
	// nplurals=2; plural=(n != 1);
	"default": func(n int) int {
		if n != 1 {
			return 1
		}

		return 0
	},
	// nplurals=6; plural=(n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 ? 4 : 5);
	"ar_AR": func(n int) int {
		if n == 0 {
			return 0
		}

		if n == 1 {
			return 1
		}

		if n == 2 {
			return 2
		}

		if n%100 >= 3 && n%100 <= 10 {
			return 3
		}

		if n%100 >= 11 {
			return 4
		}

		return 5
	},
	// nplurals=3; plural=(n==1) ? 0 : (n>=2 && n<=4) ? 1 : 2;
	"cs_CZ": func(n int) int {
		if n == 1 {
			return 0
		}

		if n >= 2 && n <= 4 {
			return 1
		}

		return 2
	},
	// nplurals=3; plural=(n==1 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2);
	"pl_PL": func(n int) int {
		if n == 1 {
			return 0
		}

		if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
			return 1
		}

		return 2
	},
	// nplurals=2; plural=(n > 1);
	"pt_BR": func(n int) int {
		if n > 1 {
			return 1
		}
		return 0
	},
	"ru_RU": pluralFormRuSrUa,
	"uk_UA": pluralFormRuSrUa,
	"sr_RS": pluralFormRuSrUa,
	// nplurals=1; plural=0;
	"zh_CN": func(n int) int {
		return 0
	},
}

// nplurals=3; plural=(n%10==1 && n%100!=11 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2);
func pluralFormRuSrUa(n int) int {
	if n%10 == 1 && n%100 != 11 {
		return 0
	}

	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return 1
	}

	return 2
}
