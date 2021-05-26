// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package date // import "miniflux.app/reader/date"

import (
	"testing"
)

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

	expectedTS := int64(1513980589)
	if date.Unix() != expectedTS {
		t.Errorf(`The Unix timestamp should be %v instead of %v`, expectedTS, date.Unix())
	}

	_, offset := date.Zone()
	expectedOffset := 0
	if offset != expectedOffset {
		t.Errorf(`The offset should be %v instead of %v`, expectedOffset, offset)
	}
}

func TestParseRSSDateGMT(t *testing.T) {
	date, err := Parse("Tue, 03 Jun 2003 09:39:21 GMT")
	if err != nil {
		t.Fatalf(`RSS dates should be parsed correctly`)
	}

	expectedTS := int64(1054633161)
	if date.Unix() != expectedTS {
		t.Errorf(`The Unix timestamp should be %v instead of %v`, expectedTS, date.Unix())
	}

	expectedLocation := "GMT"
	if date.Location().String() != expectedLocation {
		t.Errorf(`The location should be %q instead of %q`, expectedLocation, date.Location())
	}

	name, offset := date.Zone()

	expectedName := "GMT"
	if name != expectedName {
		t.Errorf(`The zone name should be %q instead of %q`, expectedName, name)
	}

	expectedOffset := 0
	if offset != expectedOffset {
		t.Errorf(`The offset should be %v instead of %v`, expectedOffset, offset)
	}
}

func TestParseRSSDatePST(t *testing.T) {
	date, err := Parse("Wed, 26 Dec 2018 10:00:54 PST")
	if err != nil {
		t.Fatalf(`RSS dates with PST timezone should be parsed correctly: %v`, err)
	}

	expectedTS := int64(1545847254)
	if date.Unix() != expectedTS {
		t.Errorf(`The Unix timestamp should be %v instead of %v`, expectedTS, date.Unix())
	}

	expectedLocation := "America/Los_Angeles"
	if date.Location().String() != expectedLocation {
		t.Errorf(`The location should be %q instead of %q`, expectedLocation, date.Location())
	}

	name, offset := date.Zone()

	expectedName := "PST"
	if name != expectedName {
		t.Errorf(`The zone name should be %q instead of %q`, expectedName, name)
	}

	expectedOffset := -28800
	if offset != expectedOffset {
		t.Errorf(`The offset should be %v instead of %v`, expectedOffset, offset)
	}
}

func TestParseRSSDateEST(t *testing.T) {
	date, err := Parse("Wed, 10 Feb 2021 22:46:00 EST")
	if err != nil {
		t.Fatalf(`RSS dates with EST timezone should be parsed correctly: %v`, err)
	}

	expectedTS := int64(1613015160)
	if date.Unix() != expectedTS {
		t.Errorf(`The Unix timestamp should be %v instead of %v`, expectedTS, date.Unix())
	}

	expectedLocation := "America/New_York"
	if date.Location().String() != expectedLocation {
		t.Errorf(`The location should be %q instead of %q`, expectedLocation, date.Location())
	}

	name, offset := date.Zone()

	expectedName := "EST"
	if name != expectedName {
		t.Errorf(`The zone name should be %q instead of %q`, expectedName, name)
	}

	expectedOffset := -18000
	if offset != expectedOffset {
		t.Errorf(`The offset should be %v instead of %v`, expectedOffset, offset)
	}
}
func TestParseRSSDateOffset(t *testing.T) {
	date, err := Parse("Sun, 28 Oct 2018 13:48:00 +0100")
	if err != nil {
		t.Fatalf(`RSS dates with offset should be parsed correctly: %v`, err)
	}

	expectedTS := int64(1540730880)
	if date.Unix() != expectedTS {
		t.Errorf(`The Unix timestamp should be %v instead of %v`, expectedTS, date.Unix())
	}

	_, offset := date.Zone()
	expectedOffset := 3600
	if offset != expectedOffset {
		t.Errorf(`The offset should be %v instead of %v`, expectedOffset, offset)
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
		"1520932969",
		"Tue 16 Feb 2016, 23:16:00 EDT",
		"Tue, 16 Feb 2016 23:16:00 EDT",
		"Tue, Feb 16 2016 23:16:00 EDT",
		"March 30 2020 07:02:38 PM",
		"Mon, 30 Mar 2020 19:53 +0000",
		"Mon, 03/30/2020 - 19:19",
		"2018-12-12T12:12",
		"2020-11-08T16:20:00-05:00Z",
		"Nov. 16, 2020, 10:57 a.m.",
		"Friday 06 November 2020",
		"Mon, November 16, 2020, 11:12 PM EST",
		"Lundi, 16. Novembre 2020 - 15:54",
		"Thu Nov 12 2020 17:00:00 GMT+0000 (Coordinated Universal Time)",
		"Sat, 11 04 2020 08:51:49 +0100",
		"Mon, 16th Nov 2020 13:16:28 GMT",
		"Nov. 2020",
		"ven., 03 juil. 2020 15:09:58 +0000",
		"Fri, 26/06/2020",
		"Thu, 29 Oct 2020 07:36:03 GMT-07:00",
		"jeu., 02 avril 2020 00:00:00 +0200",
		"Jan. 4, 2016, 12:37 p.m.",
		"2018-10-23 04:07:42 +00:00",
		"5 August, 2019",
		"mar., 01 déc. 2020 16:11:02 +0000",
		"Thurs, 15 Oct 2020 00:00:39 +0000",
		"Thur, 19 Nov 2020 00:00:39 +0000",
	}

	for _, date := range dates {
		if _, err := Parse(date); err != nil {
			t.Errorf(`Unable to parse date: %q`, date)
		}
	}
}
