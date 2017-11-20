// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale

import "log"

type Translation map[string]interface{}

type Locales map[string]Translation

func Load() *Translator {
	translator := NewTranslator()

	for language, translations := range Translations {
		log.Println("Loading translation:", language)
		translator.AddLanguage(language, translations)
	}

	return translator
}

// GetAvailableLanguages returns the list of available languages.
func GetAvailableLanguages() map[string]string {
	return map[string]string{
		"en_US": "English",
		"fr_FR": "Français",
	}
}
