// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/v2/internal/locale"

var numberOfPluralFormsPerLanguage = map[string]int{
	"en_US": 2,
	"es_ES": 2,
	"fr_FR": 2,
	"de_DE": 2,
	"pl_PL": 3,
	"pt_BR": 2,
	"zh_CN": 1,
	"zh_TW": 1,
	"nl_NL": 2,
	"ru_RU": 3,
	"it_IT": 2,
	"ja_JP": 1,
	"tr_TR": 2,
	"el_EL": 2,
	"fi_FI": 2,
	"hi_IN": 2,
	"uk_UA": 3,
	"id_ID": 1,
}

// AvailableLanguages returns the list of available languages.
func AvailableLanguages() map[string]string {
	return map[string]string{
		"en_US": "English",
		"es_ES": "Español",
		"fr_FR": "Français",
		"de_DE": "Deutsch",
		"pl_PL": "Polski",
		"pt_BR": "Português Brasileiro",
		"zh_CN": "简体中文",
		"zh_TW": "繁體中文",
		"nl_NL": "Nederlands",
		"ru_RU": "Русский",
		"it_IT": "Italiano",
		"ja_JP": "日本語",
		"tr_TR": "Türkçe",
		"el_EL": "Ελληνικά",
		"fi_FI": "Suomi",
		"hi_IN": "हिन्दी",
		"uk_UA": "Українська",
		"id_ID": "Bahasa Indonesia",
	}
}
