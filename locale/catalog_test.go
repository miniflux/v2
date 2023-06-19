// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/locale"

import "testing"

func TestParserWithInvalidData(t *testing.T) {
	_, err := parseTranslationMessages([]byte(`{`))
	if err == nil {
		t.Fatal(`An error should be returned when parsing invalid data`)
	}
}

func TestParser(t *testing.T) {
	translations, err := parseTranslationMessages([]byte(`{"k": "v"}`))
	if err != nil {
		t.Fatalf(`Unexpected parsing error: %v`, err)
	}

	if translations == nil {
		t.Fatal(`Translations should not be nil`)
	}

	value, found := translations["k"]
	if !found {
		t.Fatal(`The translation should contains the defined key`)
	}

	if value.(string) != "v" {
		t.Fatal(`The translation key should contains the defined value`)
	}
}

func TestLoadCatalog(t *testing.T) {
	if err := LoadCatalogMessages(); err != nil {
		t.Fatal(err)
	}
}

func TestAllKeysHaveValue(t *testing.T) {
	for language := range AvailableLanguages() {
		messages, err := loadTranslationFile(language)
		if err != nil {
			t.Fatalf(`Unable to load translation messages for language %q`, language)
		}

		if len(messages) == 0 {
			t.Fatalf(`The language %q doesn't have any messages`, language)
		}

		for k, v := range messages {
			switch value := v.(type) {
			case string:
				if value == "" {
					t.Errorf(`The key %q for the language %q have an empty string as value`, k, language)
				}
			case []string:
				if len(value) == 0 {
					t.Errorf(`The key %q for the language %q have an empty list as value`, k, language)
				}
			}
		}
	}
}

func TestMissingTranslations(t *testing.T) {
	refLang := "en_US"
	references, err := loadTranslationFile(refLang)
	if err != nil {
		t.Fatal(`Unable to parse reference language`)
	}

	for language := range AvailableLanguages() {
		if language == refLang {
			continue
		}

		messages, err := loadTranslationFile(language)
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
