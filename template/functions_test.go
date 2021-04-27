// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template // import "miniflux.app/template"

import (
	"testing"
	"time"

	"miniflux.app/locale"
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

func TestElapsedTime(t *testing.T) {
	printer := locale.NewPrinter("en_US")
	var dt = []struct {
		in  time.Time
		out string
	}{
		{time.Time{}, printer.Printf("time_elapsed.not_yet")},
		{time.Now().Add(time.Hour), printer.Printf("time_elapsed.not_yet")},
		{time.Now(), printer.Printf("time_elapsed.now")},
		{time.Now().Add(-time.Minute), printer.Plural("time_elapsed.minutes", 1, 1)},
		{time.Now().Add(-time.Minute * 40), printer.Plural("time_elapsed.minutes", 40, 40)},
		{time.Now().Add(-time.Hour), printer.Plural("time_elapsed.hours", 1, 1)},
		{time.Now().Add(-time.Hour * 3), printer.Plural("time_elapsed.hours", 3, 3)},
		{time.Now().Add(-time.Hour * 32), printer.Printf("time_elapsed.yesterday")},
		{time.Now().Add(-time.Hour * 24 * 3), printer.Plural("time_elapsed.days", 3, 3)},
		{time.Now().Add(-time.Hour * 24 * 14), printer.Plural("time_elapsed.days", 14, 14)},
		{time.Now().Add(-time.Hour * 24 * 15), printer.Plural("time_elapsed.days", 15, 15)},
		{time.Now().Add(-time.Hour * 24 * 21), printer.Plural("time_elapsed.weeks", 3, 3)},
		{time.Now().Add(-time.Hour * 24 * 32), printer.Plural("time_elapsed.months", 1, 1)},
		{time.Now().Add(-time.Hour * 24 * 60), printer.Plural("time_elapsed.months", 2, 2)},
		{time.Now().Add(-time.Hour * 24 * 366), printer.Plural("time_elapsed.years", 1, 1)},
		{time.Now().Add(-time.Hour * 24 * 365 * 3), printer.Plural("time_elapsed.years", 3, 3)},
	}
	for i, tt := range dt {
		if out := elapsedTime(printer, "Local", tt.in); out != tt.out {
			t.Errorf(`%d. content mismatch for "%v": expected=%q got=%q`, i, tt.in, tt.out, out)
		}
	}
}

func TestFormatFileSize(t *testing.T) {
	scenarios := []struct {
		input    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KiB"},
		{43520, "42.5 KiB"},
		{5000 * 1024 * 1024, "4.9 GiB"},
	}

	for _, scenario := range scenarios {
		result := formatFileSize(scenario.input)
		if result != scenario.expected {
			t.Errorf(`Unexpected result, got %q instead of %q for %d`, result, scenario.expected, scenario.input)
		}
	}
}

func TestBuildQuery(t *testing.T) {
	scenarios := []struct {
		input    []interface{}
		expected string
	}{
		{[]interface{}{"foo", "bar"}, "?foo=bar"},
		{[]interface{}{"foo", "bar", "foo2", "bar2"}, "?foo=bar&foo2=bar2"},
		{[]interface{}{"foo", ""}, ""},
		{[]interface{}{"foo", nil}, ""},
		{[]interface{}{"foo", 0}, ""},
		{[]interface{}{"foo", false}, ""},
		{[]interface{}{"foo", true}, "?foo=t"},
	}

	for _, scenario := range scenarios {
		result := buildQuery(scenario.input...)
		if result != scenario.expected {
			t.Errorf(`Unexpected result, got %q instead of %q for %v`, result, scenario.expected, scenario.input)
		}
	}
}
