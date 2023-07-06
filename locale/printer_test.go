// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/locale"

import "testing"

func TestTranslateWithMissingLanguage(t *testing.T) {
	defaultCatalog = catalog{}
	translation := NewPrinter("invalid").Printf("missing.key")

	if translation != "missing.key" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestTranslateWithMissingKey(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			"k": "v",
		},
	}

	translation := NewPrinter("en_US").Printf("missing.key")
	if translation != "missing.key" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestTranslateWithExistingKey(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			"auth.username": "Login",
		},
	}

	translation := NewPrinter("en_US").Printf("auth.username")
	if translation != "Login" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestTranslateWithExistingKeyAndPlaceholder(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			"key": "Test: %s",
		},
		"fr_FR": translationDict{
			"key": "Test : %s",
		},
	}

	translation := NewPrinter("fr_FR").Printf("key", "ok")
	if translation != "Test : ok" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestTranslateWithMissingKeyAndPlaceholder(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			"auth.username": "Login",
		},
		"fr_FR": translationDict{
			"auth.username": "Identifiant",
		},
	}

	translation := NewPrinter("fr_FR").Printf("Status: %s", "ok")
	if translation != "Status: ok" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestTranslateWithInvalidValue(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			"auth.username": "Login",
		},
		"fr_FR": translationDict{
			"auth.username": true,
		},
	}

	translation := NewPrinter("fr_FR").Printf("auth.username")
	if translation != "auth.username" {
		t.Errorf(`Wrong translation, got %q`, translation)
	}
}

func TestTranslatePluralWithDefaultRule(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			"number_of_users": []string{"%d user (%s)", "%d users (%s)"},
		},
		"fr_FR": translationDict{
			"number_of_users": []string{"%d utilisateur (%s)", "%d utilisateurs (%s)"},
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

func TestTranslatePluralWithRussianRule(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			"time_elapsed.years": []string{"%d year", "%d years"},
		},
		"ru_RU": translationDict{
			"time_elapsed.years": []string{"%d год назад", "%d года назад", "%d лет назад"},
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

func TestTranslatePluralWithMissingTranslation(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			"number_of_users": []string{"%d user (%s)", "%d users (%s)"},
		},
		"fr_FR": translationDict{},
	}
	translation := NewPrinter("fr_FR").Plural("number_of_users", 2)
	expected := "number_of_users"
	if translation != expected {
		t.Errorf(`Wrong translation, got %q instead of %q`, translation, expected)
	}
}

func TestTranslatePluralWithInvalidValues(t *testing.T) {
	defaultCatalog = catalog{
		"en_US": translationDict{
			"number_of_users": []string{"%d user (%s)", "%d users (%s)"},
		},
		"fr_FR": translationDict{
			"number_of_users": "must be a slice",
		},
	}
	translation := NewPrinter("fr_FR").Plural("number_of_users", 2)
	expected := "number_of_users"
	if translation != expected {
		t.Errorf(`Wrong translation, got %q instead of %q`, translation, expected)
	}
}
