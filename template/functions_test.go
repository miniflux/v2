// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template // import "miniflux.app/template"

import (
	"testing"
)

func TestDict(t *testing.T) {
	d, err := dict("k1", "v1", "k2", "v2")
	if err != nil {
		t.Fatalf(`The dict should be valid: %v`, err)
	}

	if value, found := d["k1"]; found {
		if value != "v1" {
			t.Fatalf(`Unexpected value for k1: got %q`, value)
		}
	}

	if value, found := d["k2"]; found {
		if value != "v2" {
			t.Fatalf(`Unexpected value for k2: got %q`, value)
		}
	}
}

func TestDictWithInvalidNumberOfArguments(t *testing.T) {
	_, err := dict("k1")
	if err == nil {
		t.Fatal(`An error should be returned if the number of arguments are not even`)
	}
}

func TestDictWithInvalidMap(t *testing.T) {
	_, err := dict(1, 2)
	if err == nil {
		t.Fatal(`An error should be returned if the dict keys are not string`)
	}
}

func TestHasKey(t *testing.T) {
	input := map[string]string{"k": "v"}

	if !hasKey(input, "k") {
		t.Fatal(`This key exists in the map and should returns true`)
	}

	if hasKey(input, "missing") {
		t.Fatal(`This key doesn't exists in the given map and should returns false`)
	}
}

func TestTruncateWithShortTexts(t *testing.T) {
	scenarios := []string{"Short text", "Короткий текст"}

	for _, input := range scenarios {
		result := truncate(input, 25)
		if result != input {
			t.Fatalf(`Unexpected output, got %q instead of %q`, result, input)
		}

		result = truncate(input, len(input))
		if result != input {
			t.Fatalf(`Unexpected output, got %q instead of %q`, result, input)
		}
	}
}

func TestTruncateWithLongTexts(t *testing.T) {
	scenarios := map[string]string{
		"This is a really pretty long English text": "This is a really pretty l…",
		"Это реально очень длинный русский текст":   "Это реально очень длинный…",
	}

	for input, expected := range scenarios {
		result := truncate(input, 25)
		if result != expected {
			t.Fatalf(`Unexpected output, got %q instead of %q`, result, expected)
		}
	}
}

func TestIsEmail(t *testing.T) {
	if !isEmail("user@domain.tld") {
		t.Fatal(`This email is valid and should returns true`)
	}

	if isEmail("invalid") {
		t.Fatal(`This email is not valid and should returns false`)
	}
}
