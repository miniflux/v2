// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Translator manage supported locales.
type Translator struct {
	locales Locales
}

// AddLanguage loads a new language into the system.
func (t *Translator) AddLanguage(language, translations string) error {
	var decodedTranslations Translation

	decoder := json.NewDecoder(strings.NewReader(translations))
	if err := decoder.Decode(&decodedTranslations); err != nil {
		return fmt.Errorf("Invalid JSON file: %v", err)
	}

	t.locales[language] = decodedTranslations
	return nil
}

// GetLanguage returns the given language handler.
func (t *Translator) GetLanguage(language string) *Language {
	translations, found := t.locales[language]
	if !found {
		return &Language{language: language}
	}

	return &Language{language: language, translations: translations}
}

// NewTranslator creates a new Translator.
func NewTranslator() *Translator {
	return &Translator{locales: make(Locales)}
}
