// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

import "testing"

func TestParserWithInvalidData(t *testing.T) {
	_, err := parseTranslationDict(`{`)
	if err == nil {
		t.Fatal(`An error should be returned when parsing invalid data`)
	}
}

func TestParser(t *testing.T) {
	translations, err := parseTranslationDict(`{"k": "v"}`)
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
