// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Translator struct {
	Locales Locales
}

func (t *Translator) AddLanguage(language, translations string) error {
	var decodedTranslations Translation

	decoder := json.NewDecoder(strings.NewReader(translations))
	if err := decoder.Decode(&decodedTranslations); err != nil {
		return fmt.Errorf("Invalid JSON file: %v", err)
	}

	t.Locales[language] = decodedTranslations
	return nil
}

func (t *Translator) GetLanguage(language string) *Language {
	translations, found := t.Locales[language]
	if !found {
		return &Language{language: language}
	}

	return &Language{language: language, translations: translations}
}

func NewTranslator() *Translator {
	return &Translator{Locales: make(Locales)}
}
