// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/v2/internal/locale"

import "testing"

func TestPrintfWithMissingLanguage(t *testing.T) {
	defaultCatalog = catalog{}
	translation := NewPrinter("invalid").Printf("missing.key")

	if translation != "missing.key" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestPrintfWithMissingKey(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			singulars: map[string]string{
				"k": "v",
			},
		},
	}

	translation := NewPrinter("en_US").Printf("missing.key")
	if translation != "missing.key" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestPrintfWithExistingKey(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			singulars: map[string]string{
				"auth.username": "Login",
			},
		},
	}

	translation := NewPrinter("en_US").Printf("auth.username")
	if translation != "Login" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestPrintfWithExistingKeyAndPlaceholder(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			singulars: map[string]string{
				"key": "Test: %s",
			},
		},
		"fr_FR": translationDict{
			singulars: map[string]string{
				"key": "Test : %s",
			},
		},
	}

	translation := NewPrinter("fr_FR").Printf("key", "ok")
	if translation != "Test : ok" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestPrintfWithMissingKeyAndPlaceholder(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			singulars: map[string]string{
				"auth.username": "Login",
			},
		},
		"fr_FR": translationDict{
			singulars: map[string]string{
				"auth.username": "Identifiant",
			},
		},
	}

	translation := NewPrinter("fr_FR").Printf("Status: %s", "ok")
	if translation != "Status: ok" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestPrintWithMissingLanguage(t *testing.T) {
	defaultCatalog = catalog{}
	translation := NewPrinter("invalid").Print("missing.key")

	if translation != "missing.key" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestPrintWithMissingKey(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			singulars: map[string]string{
				"existing.key": "value",
			},
		},
	}

	translation := NewPrinter("en_US").Print("missing.key")
	if translation != "missing.key" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestPrintWithExistingKey(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			singulars: map[string]string{
				"auth.username": "Login",
			},
		},
	}

	translation := NewPrinter("en_US").Print("auth.username")
	if translation != "Login" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestPrintWithDifferentLanguages(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			singulars: map[string]string{
				"greeting": "Hello",
			},
		},
		"fr_FR": translationDict{
			singulars: map[string]string{
				"greeting": "Bonjour",
			},
		},
		"es_ES": translationDict{
			singulars: map[string]string{
				"greeting": "Hola",
			},
		},
	}

	tests := []struct {
		language string
		expected string
	}{
		{"en_US", "Hello"},
		{"fr_FR", "Bonjour"},
		{"es_ES", "Hola"},
	}

	for _, test := range tests {
		translation := NewPrinter(test.language).Print("greeting")
		if translation != test.expected {
			t.Errorf(`Wrong translation for %s, got %q instead of %q`, test.language, translation, test.expected)
		}
	}
}

func TestPrintWithEmptyKey(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			singulars: map[string]string{
				"": "empty key translation",
			},
		},
	}

	translation := NewPrinter("en_US").Print("")
	if translation != "empty key translation" {
		t.Errorf(`Wrong translation for empty key, got %q`, translation)
	}
}

func TestPrintWithEmptyTranslation(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			singulars: map[string]string{
				"empty.value": "",
			},
		},
	}

	translation := NewPrinter("en_US").Print("empty.value")
	if translation != "" {
		t.Errorf(`Wrong translation for empty value, got %q`, translation)
	}
}

func TestPluralWithDefaultRule(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			plurals: map[string][]string{
				"number_of_users": {"%d user (%s)", "%d users (%s)"},
			},
		},
		"fr_FR": translationDict{
			plurals: map[string][]string{
				"number_of_users": {"%d utilisateur (%s)", "%d utilisateurs (%s)"},
			},
		},
	}

	printer := NewPrinter("fr_FR")
	translation := printer.Plural("number_of_users", 1, 1, "some text")
	expected := "1 utilisateur (some text)"
	if translation != expected {
		t.Errorf(`Wrong translation, got %q instead of %q`, translation, expected)
	}

	translation = printer.Plural("number_of_users", 2, 2, "some text")
	expected = "2 utilisateurs (some text)"
	if translation != expected {
		t.Errorf(`Wrong translation, got %q instead of %q`, translation, expected)
	}
}

func TestPluralWithRussianRule(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			plurals: map[string][]string{
				"time_elapsed.years": {"%d year", "%d years"},
			},
		},
		"ru_RU": translationDict{
			plurals: map[string][]string{
				"time_elapsed.years": {"%d год назад", "%d года назад", "%d лет назад"},
			},
		},
	}

	printer := NewPrinter("ru_RU")

	translation := printer.Plural("time_elapsed.years", 1, 1)
	expected := "1 год назад"
	if translation != expected {
		t.Errorf(`Wrong translation, got %q instead of %q`, translation, expected)
	}

	translation = printer.Plural("time_elapsed.years", 2, 2)
	expected = "2 года назад"
	if translation != expected {
		t.Errorf(`Wrong translation, got %q instead of %q`, translation, expected)
	}

	translation = printer.Plural("time_elapsed.years", 5, 5)
	expected = "5 лет назад"
	if translation != expected {
		t.Errorf(`Wrong translation, got %q instead of %q`, translation, expected)
	}
}

func TestPluralWithMissingTranslation(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			plurals: map[string][]string{
				"number_of_users": {"%d user (%s)", "%d users (%s)"},
			},
		},
		"fr_FR": translationDict{},
	}
	translation := NewPrinter("fr_FR").Plural("number_of_users", 2)
	expected := "number_of_users"
	if translation != expected {
		t.Errorf(`Wrong translation, got %q instead of %q`, translation, expected)
	}
}

func TestPluralWithMissingLanguage(t *testing.T) {
	defaultCatalog = catalog{}
	translation := NewPrinter("invalid_language").Plural("test.key", 2)
	expected := "test.key"
	if translation != expected {
		t.Errorf(`Wrong translation, got %q instead of %q`, translation, expected)
	}
}

func TestPluralWithIndexOutOfBounds(t *testing.T) {
	defaultCatalog = catalog{
		"test_lang": translationDict{
			plurals: map[string][]string{
				"limited.key": {"only one form"},
			},
		},
	}

	// Force a scenario where getPluralForm might return an index >= len(plurals)
	// We'll create a scenario with Czech language rules
	defaultCatalog["cs_CZ"] = translationDict{
		plurals: map[string][]string{
			"limited.key": {"one form only"}, // Only one form, but Czech has 3 plural forms
		},
	}

	printer := NewPrinter("cs_CZ")
	// n=5 should return index 2 for Czech, but we only have 1 form (index 0)
	translation := printer.Plural("limited.key", 5)
	expected := "limited.key"
	if translation != expected {
		t.Errorf(`Wrong translation for out of bounds index, got %q instead of %q`, translation, expected)
	}
}

func TestPluralWithVariousLanguageRules(t *testing.T) {
	defaultCatalog = catalog{
		"ar_AR": translationDict{
			plurals: map[string][]string{
				"items": {"no items", "one item", "two items", "few items", "many items", "other items"},
			},
		},
		"pl_PL": translationDict{
			plurals: map[string][]string{
				"files": {"one file", "few files", "many files"},
			},
		},
		"ja_JP": translationDict{
			plurals: map[string][]string{
				"photos": {"photos"},
			},
		},
	}

	tests := []struct {
		language string
		key      string
		n        int
		expected string
	}{
		// Arabic tests
		{"ar_AR", "items", 0, "no items"},
		{"ar_AR", "items", 1, "one item"},
		{"ar_AR", "items", 2, "two items"},
		{"ar_AR", "items", 5, "few items"},   // n%100 >= 3 && n%100 <= 10
		{"ar_AR", "items", 15, "many items"}, // n%100 >= 11

		// Polish tests
		{"pl_PL", "files", 1, "one file"},
		{"pl_PL", "files", 3, "few files"},  // n%10 >= 2 && n%10 <= 4
		{"pl_PL", "files", 5, "many files"}, // default case

		// Japanese tests (always uses same form)
		{"ja_JP", "photos", 1, "photos"},
		{"ja_JP", "photos", 10, "photos"},
	}

	for _, test := range tests {
		printer := NewPrinter(test.language)
		translation := printer.Plural(test.key, test.n)
		if translation != test.expected {
			t.Errorf(`Wrong translation for %s with n=%d, got %q instead of %q`,
				test.language, test.n, translation, test.expected)
		}
	}
}
