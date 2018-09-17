// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

import "miniflux.app/logger"

// Translation is the translation mapping table.
type Translation map[string]interface{}

// Locales represents locales supported by the system.
type Locales map[string]Translation

// Load prepare the locale system by loading all translations.
func Load() *Translator {
	translator := NewTranslator()

	for language, tr := range translations {
		logger.Debug("Loading translation: %s", language)
		translator.AddLanguage(language, tr)
	}

	return translator
}

// AvailableLanguages returns the list of available languages.
func AvailableLanguages() map[string]string {
	return map[string]string{
		"en_US": "English",
		"fr_FR": "Français",
		"de_DE": "Deutsch",
		"pl_PL": "Polski",
		"zh_CN": "简体中文",
		"nl_NL": "Nederlands",
		"ru_RU": "Русский",
	}
}
