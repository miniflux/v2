// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package date

import "testing"

func TestParseEmptyDate(t *testing.T) {
	if _, err := Parse("  "); err == nil {
		t.Fatalf(`Empty dates should return an error`)
	}
}

func TestParseInvalidDate(t *testing.T) {
	if _, err := Parse("invalid"); err == nil {
		t.Fatalf(`Invalid dates should return an error`)
	}
}

func TestParseAtomDate(t *testing.T) {
	date, err := Parse("2017-12-22T22:09:49+00:00")
	if err != nil {
		t.Fatalf(`Atom dates should be parsed correctly`)
	}

	if date.Unix() != 1513980589 {
		t.Fatal(`Invalid date parsed`)
	}
}

func TestParseRSSDate(t *testing.T) {
	date, err := Parse("Tue, 03 Jun 2003 09:39:21 GMT")
	if err != nil {
		t.Fatalf(`RSS dates should be parsed correctly`)
	}

	if date.Unix() != 1054633161 {
		t.Fatal(`Invalid date parsed`)
	}
}

func TestParseWeirdDateFormat(t *testing.T) {
	dates := []string{
		"Sun, 17 Dec 2017 1:55 PM EST",
		"9 Dec 2016 12:00 GMT",
		"Friday, December 22, 2017 - 3:09pm",
		"Friday, December 8, 2017 - 3:07pm",
		"Thu, 25 Feb 2016 00:00:00 Europe/Brussels",
		"Mon, 09 Apr 2018, 16:04",
		"Di, 23 Jan 2018 00:00:00 +0100",
		"Do, 29 Mär 2018 00:00:00 +0200",
		"mer, 9 avr 2018 00:00:00 +0200",
	}

	for _, date := range dates {
		if _, err := Parse(date); err != nil {
			t.Fatalf(`Unable to parse date: %q`, date)
		}
	}
}
