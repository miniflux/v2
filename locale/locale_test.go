// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.
package locale // import "miniflux.app/locale"

import "testing"

func TestTranslateWithMissingLanguage(t *testing.T) {
	translator := NewTranslator()
	translation := translator.GetLanguage("en_US").Get("auth.username")

	if translation != "auth.username" {
		t.Errorf("Wrong translation, got %s", translation)
	}
}

func TestTranslateWithExistingKey(t *testing.T) {
	data := `{"auth.username": "Username"}`
	translator := NewTranslator()
	translator.AddLanguage("en_US", data)
	translation := translator.GetLanguage("en_US").Get("auth.username")

	if translation != "Username" {
		t.Errorf("Wrong translation, got %s", translation)
	}
}

func TestTranslateWithMissingKey(t *testing.T) {
	data := `{"auth.username": "Username"}`
	translator := NewTranslator()
	translator.AddLanguage("en_US", data)
	translation := translator.GetLanguage("en_US").Get("auth.password")

	if translation != "auth.password" {
		t.Errorf("Wrong translation, got %s", translation)
	}
}

func TestTranslateWithMissingKeyAndPlaceholder(t *testing.T) {
	translator := NewTranslator()
	translator.AddLanguage("fr_FR", "")
	translation := translator.GetLanguage("fr_FR").Get("Status: %s", "ok")

	if translation != "Status: ok" {
		t.Errorf("Wrong translation, got %s", translation)
	}
}

func TestTranslatePluralWithDefaultRule(t *testing.T) {
	data := `{"number_of_users": ["Il y a %d utilisateur (%s)", "Il y a %d utilisateurs (%s)"]}`
	translator := NewTranslator()
	translator.AddLanguage("fr_FR", data)
	language := translator.GetLanguage("fr_FR")

	translation := language.Plural("number_of_users", 1, 1, "some text")
	expected := "Il y a 1 utilisateur (some text)"
	if translation != expected {
		t.Errorf(`Wrong translation, got "%s" instead of "%s"`, translation, expected)
	}

	translation = language.Plural("number_of_users", 2, 2, "some text")
	expected = "Il y a 2 utilisateurs (some text)"
	if translation != expected {
		t.Errorf(`Wrong translation, got "%s" instead of "%s"`, translation, expected)
	}
}

func TestTranslatePluralWithRussianRule(t *testing.T) {
	data := `{"key": ["из %d книги за %d день", "из %d книг за %d дня", "из %d книг за %d дней"]}`
	translator := NewTranslator()
	translator.AddLanguage("ru_RU", data)
	language := translator.GetLanguage("ru_RU")

	translation := language.Plural("key", 1, 1, 1)
	expected := "из 1 книги за 1 день"
	if translation != expected {
		t.Errorf(`Wrong translation, got "%s" instead of "%s"`, translation, expected)
	}

	translation = language.Plural("key", 2, 2, 2)
	expected = "из 2 книг за 2 дня"
	if translation != expected {
		t.Errorf(`Wrong translation, got "%s" instead of "%s"`, translation, expected)
	}

	translation = language.Plural("key", 5, 5, 5)
	expected = "из 5 книг за 5 дней"
	if translation != expected {
		t.Errorf(`Wrong translation, got "%s" instead of "%s"`, translation, expected)
	}
}

func TestTranslatePluralWithMissingTranslation(t *testing.T) {
	translator := NewTranslator()
	translator.AddLanguage("fr_FR", "")
	language := translator.GetLanguage("fr_FR")

	translation := language.Plural("number_of_users", 2)
	expected := "number_of_users"
	if translation != expected {
		t.Errorf(`Wrong translation, got "%s" instead of "%s"`, translation, expected)
	}
}
