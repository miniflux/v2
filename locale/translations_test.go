// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

import "testing"

func TestAllLanguagesHaveCatalog(t *testing.T) {
	for language := range AvailableLanguages() {
		if _, found := translations[language]; !found {
			t.Fatalf(`This language do not have a catalog: %q`, language)
		}
	}
}

func TestAllKeysHaveValue(t *testing.T) {
	for language := range AvailableLanguages() {
		messages, err := parseTranslationDict(translations[language])
		if err != nil {
			t.Fatalf(`Parsing error for language %q`, language)
		}

		if len(messages) == 0 {
			t.Fatalf(`The language %q doesn't have any messages`, language)
		}

		for k, v := range messages {
			switch value := v.(type) {
			case string:
				if value == "" {
					t.Fatalf(`The key %q for the language %q have an empty string as value`, k, language)
				}
			case []string:
				if len(value) == 0 {
					t.Fatalf(`The key %q for the language %q have an empty list as value`, k, language)
				}
			}
		}
	}
}

func TestMissingTranslations(t *testing.T) {
	refLang := "en_US"
	references, err := parseTranslationDict(translations[refLang])
	if err != nil {
		t.Fatal(`Unable to parse reference language`)
	}

	for language := range AvailableLanguages() {
		if language == refLang {
			continue
		}

		messages, err := parseTranslationDict(translations[language])
		if err != nil {
			t.Fatalf(`Parsing error for language %q`, language)
		}

		for key := range references {
			if _, found := messages[key]; !found {
				t.Fatalf(`Translation key %q not found in language %q`, key, language)
			}
		}
	}
}
