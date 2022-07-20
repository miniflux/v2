// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

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
	}
}
