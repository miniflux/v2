// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package date // import "miniflux.app/v2/internal/reader/date"

import (
	"strings"
	"testing"
	"time"
)

// validDates lists inputs that Parse must accept, with the instant each one
// represents, written in RFC 3339 with the offset given by the input.
var validDates = []struct {
	input string
	want  string
}{
	// Standard formats.
	{"2017-12-22T22:09:49+00:00", "2017-12-22T22:09:49Z"},
	{"1985-04-12T23:20:50.52Z", "1985-04-12T23:20:50.52Z"},
	{"03 Jun 03 09:39 GMT", "2003-06-03T09:39:00Z"},
	{"03 Jun 03 09:39 -0700", "2003-06-03T09:39:00-07:00"},
	{"Tuesday, 03-Jun-03 09:39:21 GMT", "2003-06-03T09:39:21Z"},
	{"Tue, 03 Jun 2003 09:39:21 GMT", "2003-06-03T09:39:21Z"},
	{"Sun, 28 Oct 2018 13:48:00 +0100", "2018-10-28T13:48:00+01:00"},
	{"Wed Feb 10 22:46:00 UTC 2021", "2021-02-10T22:46:00Z"},
	{"Wed Feb 10 22:46:00 +0100 2021", "2021-02-10T22:46:00+01:00"},
	{"Wed Feb 10 22:46:00 2021", "2021-02-10T22:46:00Z"},

	// RFC 3339 section 5.6 allows "T" and "Z" to be lower case.
	{"2003-12-13t18:30:02z", "2003-12-13T18:30:02Z"},
	{"2003-12-13T18:30:02z", "2003-12-13T18:30:02Z"},
	{"2003-12-13t18:30:02Z", "2003-12-13T18:30:02Z"},
	{"1985-04-12t23:20:50.52z", "1985-04-12T23:20:50.52Z"},
	{"1996-12-19t16:39:57-08:00", "1996-12-19T16:39:57-08:00"},
	// The separators are looked for at fixed positions, after the input is trimmed.
	{" 2003-12-13t18:30:02z\n", "2003-12-13T18:30:02Z"},

	// Unix timestamps.
	{"1520932969", "2018-03-13T09:22:49Z"},
	{"0", "1970-01-01T00:00:00Z"},
	{"-1", "1969-12-31T23:59:59Z"},

	// Time zone abbreviations and names.
	{"Wed, 26 Dec 2018 10:00:54 PST", "2018-12-26T10:00:54-08:00"},
	{"Tue, 26 Jun 2018 10:00:54 PDT", "2018-06-26T10:00:54-07:00"},
	{"Wed, 10 Feb 2021 22:46:00 EST", "2021-02-10T22:46:00-05:00"},
	{"Thu, 10 Jun 2021 22:46:00 EDT", "2021-06-10T22:46:00-04:00"},
	{"Fri, 31 Mar 2023 20:19:00 America/Los_Angeles", "2023-03-31T20:19:00-07:00"},
	{"Tue, 31 Jan 2023 20:19:00 America/Los_Angeles", "2023-01-31T20:19:00-08:00"},
	{"Thu, 25 Feb 2016 00:00:00 Europe/Brussels", "2016-02-25T00:00:00+01:00"},
	{"Mon, 25 Jul 2016 00:00:00 Europe/Brussels", "2016-07-25T00:00:00+02:00"},

	// Non-standard formats found in feeds.
	{"Sun, 17 Dec 2017 1:55 PM EST", "2017-12-17T13:55:00-05:00"},
	{"9 Dec 2016 12:00 GMT", "2016-12-09T12:00:00Z"},
	{"Friday, December 22, 2017 - 3:09pm", "2017-12-22T15:09:00Z"},
	{"Friday, December 8, 2017 - 3:07pm", "2017-12-08T15:07:00Z"},
	{"Mon, 09 Apr 2018, 16:04", "2018-04-09T16:04:00Z"},
	{"Tue 16 Feb 2016, 23:16:00 EDT", "2016-02-16T23:16:00-04:00"},
	{"Tue, 16 Feb 2016 23:16:00 EDT", "2016-02-16T23:16:00-04:00"},
	{"Tue, Feb 16 2016 23:16:00 EDT", "2016-02-16T23:16:00-04:00"},
	{"March 30 2020 07:02:38 PM", "2020-03-30T19:02:38Z"},
	{"Mon, 30 Mar 2020 19:53 +0000", "2020-03-30T19:53:00Z"},
	{"Mon, 03/30/2020 - 19:19", "2020-03-30T19:19:00Z"},
	{"2018-12-12T12:12", "2018-12-12T12:12:00Z"},
	{"2020-11-08T16:20:00-05:00Z", "2020-11-08T16:20:00-05:00"},
	{"Friday 06 November 2020", "2020-11-06T00:00:00Z"},
	{"Mon, November 16, 2020, 11:12 PM EST", "2020-11-16T23:12:00-05:00"},
	{"Thu Nov 12 2020 17:00:00 GMT+0000 (Coordinated Universal Time)", "2020-11-12T17:00:00Z"},
	{"Sat, 11 04 2020 08:51:49 +0100", "2020-04-11T08:51:49+01:00"},
	{"Mon, 16th Nov 2020 13:16:28 GMT", "2020-11-16T13:16:28Z"},
	{"Nov. 2020", "2020-11-01T00:00:00Z"},
	{"Fri, 26/06/2020", "2020-06-26T00:00:00Z"},
	{"Thu, 29 Oct 2020 07:36:03 GMT-07:00", "2020-10-29T07:36:03-07:00"},
	{"2018-10-23 04:07:42 +00:00", "2018-10-23T04:07:42Z"},
	{"5 August, 2019", "2019-08-05T00:00:00Z"},
	{"Thurs, 15 Oct 2020 00:00:39 +0000", "2020-10-15T00:00:39Z"},
	{"Thur, 19 Nov 2020 00:00:39 +0000", "2020-11-19T00:00:39Z"},
	{"26 Sep 2022 GMT", "2022-09-26T00:00:00Z"},
	{"Thu, June 22, 2023 at 01:11 PM EDT", "2023-06-22T13:11:00-04:00"},
	{"Apr 16, 2023 08:01 GMT", "2023-04-16T08:01:00Z"},
	{"Jun 23, 2023 19:00 GMT", "2023-06-23T19:00:00Z"},
	{"09/15/2014 4:20 pm PST", "2014-09-15T16:20:00-08:00"},
	{"Fri, 23rd Jun 2023 09:32:20 GMT", "2023-06-23T09:32:20Z"},
	{"Sat, Oct 28 2023 08:28:28 PM", "2023-10-28T20:28:28Z"},
	{"Monday, October 6, 2023 - 16:29\n", "2023-10-06T16:29:00Z"},
	{"10/30/23 21:55:58", "2023-10-30T21:55:58Z"},
	{"30.10.23", "2023-10-30T00:00:00Z"},
	{"Mon,02 March 2020 10:30:00 GMT", "2020-03-02T10:30:00Z"},
	{"tue,2 Jan 2018 10:00:00 GMT", "2018-01-02T10:00:00Z"},

	// Three-digit offsets.
	{"Tue, Feb 16 2016 23:16:00 -700", "2016-02-16T23:16:00-07:00"},
	{"Sun, 28 Oct 2018 13:48:00 +100", "2018-10-28T13:48:00+01:00"},

	// Spelled out a.m. and p.m.
	{"Nov. 16, 2020, 10:57 a.m.", "2020-11-16T10:57:00Z"},
	{"Jan. 4, 2016, 12:37 p.m.", "2016-01-04T12:37:00Z"},
	{"Jan. 4, 2016, 1:37 p.m.", "2016-01-04T13:37:00Z"},
	{"Jan. 4, 2016, 01:37 p.m.", "2016-01-04T13:37:00Z"},
	{"Jan. 4, 2016, 12:37 a.m.", "2016-01-04T00:37:00Z"},
	{"November 16, 2020, 3:04 a.m.", "2020-11-16T03:04:00Z"},
	{"November 16, 2020, 3:04 p.m.", "2020-11-16T15:04:00Z"},

	// German dates.
	{"Di, 23 Jan 2018 00:00:00 +0100", "2018-01-23T00:00:00+01:00"},
	{"Do, 29 Mär 2018 00:00:00 +0200", "2018-03-29T00:00:00+02:00"},
	{"Mo, 02 Mai 2022 10:00:00 +0200", "2022-05-02T10:00:00+02:00"},
	{"Fr, 01 Okt 2021 10:00:00 +0200", "2021-10-01T10:00:00+02:00"},
	{"So, 05 Dez 2021 10:00:00 +0100", "2021-12-05T10:00:00+01:00"},

	// French dates.
	{"mer, 9 avr 2018 00:00:00 +0200", "2018-04-09T00:00:00+02:00"},
	{"ven., 03 juil. 2020 15:09:58 +0000", "2020-07-03T15:09:58Z"},
	{"jeu., 02 avril 2020 00:00:00 +0200", "2020-04-02T00:00:00+02:00"},
	{"mar., 01 déc. 2020 16:11:02 +0000", "2020-12-01T16:11:02Z"},
	{"mer., 08 janvier 2020 10:00:00 +0100", "2020-01-08T10:00:00+01:00"},
	{"sam., 08 février 2020 10:00:00 +0100", "2020-02-08T10:00:00+01:00"},
	{"dim., 08 mars 2020 10:00:00 +0100", "2020-03-08T10:00:00+01:00"},
	{"mer., 08 avril 2020 10:00:00 +0200", "2020-04-08T10:00:00+02:00"},
	{"ven., 08 mai 2020 10:00:00 +0200", "2020-05-08T10:00:00+02:00"},
	{"lun., 08 juin 2020 10:00:00 +0200", "2020-06-08T10:00:00+02:00"},
	{"mer., 08 juillet 2020 10:00:00 +0200", "2020-07-08T10:00:00+02:00"},
	{"sam., 08 août 2020 10:00:00 +0200", "2020-08-08T10:00:00+02:00"},
	{"mar., 08 septembre 2020 10:00:00 +0200", "2020-09-08T10:00:00+02:00"},
	{"jeu., 08 octobre 2020 10:00:00 +0200", "2020-10-08T10:00:00+02:00"},
	{"dim., 08 novembre 2020 10:00:00 +0100", "2020-11-08T10:00:00+01:00"},
	{"mar., 08 décembre 2020 10:00:00 +0100", "2020-12-08T10:00:00+01:00"},
	{"Lundi, 16. Novembre 2020 - 15:54", "2020-11-16T15:54:00Z"},
	{"Samedi, 05 Juillet 2025 10:00:00 +0200", "2025-07-05T10:00:00+02:00"},
	{"Dimanche, 06 décembre 2020 10:00:00 +0100", "2020-12-06T10:00:00+01:00"},
}

func mustParseRFC3339(tb testing.TB, value string) time.Time {
	tb.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		tb.Fatalf("Unable to parse RFC 3339 date %q: %v", value, err)
	}

	return parsed
}

func loadLocation(tb testing.TB, name string) *time.Location {
	tb.Helper()

	location, err := time.LoadLocation(name)
	if err != nil {
		tb.Fatalf("Unable to load location %q: %v", name, err)
	}

	return location
}

// setLocalTimezone replaces time.Local until the end of the test. Tests that
// call it must not run in parallel.
func setLocalTimezone(t *testing.T, location *time.Location) {
	t.Helper()

	previous := time.Local
	time.Local = location
	t.Cleanup(func() { time.Local = previous })
}

func FuzzParse(f *testing.F) {
	for _, testCase := range validDates {
		f.Add(testCase.input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		parsed, err := Parse(input)
		if err != nil {
			return
		}

		if _, offset := parsed.Zone(); offset > 14*60*60 || offset < -12*60*60 {
			t.Errorf("Parse(%q) returned the out of range offset %d", input, offset)
		}
	})
}

func TestParse(t *testing.T) {
	for _, testCase := range validDates {
		t.Run(testCase.input, func(t *testing.T) {
			want := mustParseRFC3339(t, testCase.want)

			got, err := Parse(testCase.input)
			if err != nil {
				t.Fatalf("Unable to parse date %q: %v", testCase.input, err)
			}

			if !got.Equal(want) {
				t.Errorf("Parse(%q) = %s, want %s", testCase.input, got.Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
			}
		})
	}
}

func TestParseInvalidDates(t *testing.T) {
	inputs := []string{
		"",
		"  ",
		"\t\n",
		"invalid",
		"2023-13-01",
		"Mon, 32 Jan 2006 15:04:05 GMT",
		"99999999999999999999",
		// No year.
		"Tue, 16 Feb 23:16:00 GMT",
		// One-digit year.
		"6/1/2 15:04",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			parsed, err := Parse(input)
			if err == nil {
				t.Fatalf("Parse(%q) = %s, want an error", input, parsed.Format(time.RFC3339Nano))
			}

			if trimmedInput := strings.TrimSpace(input); trimmedInput != "" && !strings.Contains(err.Error(), trimmedInput) {
				t.Errorf("The error %q should contain the input %q", err, trimmedInput)
			}
		})
	}
}

func TestLayoutsAreUnique(t *testing.T) {
	seen := make(map[string]bool, len(layouts))
	for _, layout := range layouts {
		if seen[layout] {
			t.Errorf("The layout %q is listed more than once", layout)
		}
		seen[layout] = true
	}
}

func TestParseTimezone(t *testing.T) {
	testCases := []struct {
		input string
		// wantLocation and wantZone are not checked when empty.
		wantLocation string
		wantZone     string
		wantOffset   int
	}{
		{"Tue, 03 Jun 2003 09:39:21 GMT", "GMT", "GMT", 0},
		{"Wed, 26 Dec 2018 10:00:54 PST", "America/Los_Angeles", "PST", -8 * 60 * 60},
		{"Tue, 26 Jun 2018 10:00:54 PDT", "America/Los_Angeles", "PDT", -7 * 60 * 60},
		{"Wed, 10 Feb 2021 22:46:00 EST", "America/New_York", "EST", -5 * 60 * 60},
		{"Thu, 10 Jun 2021 22:46:00 EDT", "America/New_York", "EDT", -4 * 60 * 60},
		{"Mon, November 16, 2020, 11:12 PM EST", "America/New_York", "EST", -5 * 60 * 60},
		{"Fri, 31 Mar 2023 20:19:00 America/Los_Angeles", "America/Los_Angeles", "PDT", -7 * 60 * 60},
		{"Tue, 31 Jan 2023 20:19:00 America/Los_Angeles", "America/Los_Angeles", "PST", -8 * 60 * 60},
		{"Thu, 25 Feb 2016 00:00:00 Europe/Brussels", "Europe/Brussels", "CET", 1 * 60 * 60},
		{"Mon, 25 Jul 2016 00:00:00 Europe/Brussels", "Europe/Brussels", "CEST", 2 * 60 * 60},
		{"Sun, 28 Oct 2018 13:48:00 +0100", "", "", 1 * 60 * 60},
		{"2017-12-22T22:09:49+00:00", "", "", 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			parsed, err := Parse(testCase.input)
			if err != nil {
				t.Fatalf("Unable to parse date %q: %v", testCase.input, err)
			}

			if testCase.wantLocation != "" && parsed.Location().String() != testCase.wantLocation {
				t.Errorf("The location should be %q instead of %q", testCase.wantLocation, parsed.Location())
			}

			zone, offset := parsed.Zone()
			if testCase.wantZone != "" && zone != testCase.wantZone {
				t.Errorf("The zone name should be %q instead of %q", testCase.wantZone, zone)
			}

			if offset != testCase.wantOffset {
				t.Errorf("The offset should be %d instead of %d", testCase.wantOffset, offset)
			}
		})
	}
}

func TestParseDoesNotDependOnLocalTimezone(t *testing.T) {
	// Go resolves a time zone abbreviation with the location the date is parsed
	// in, which is time.Local for time.Parse.
	inputs := []string{
		"Sun, 17 Dec 2017 1:55 PM EST",
		"Tue 16 Feb 2016, 23:16:00 EDT",
		"Thu, June 22, 2023 at 01:11 PM EDT",
		"09/15/2014 4:20 pm PST",
		"Thu, 25 Feb 2016 00:00 CET",
		"Fri, 31 Mar 2023 20:19:00 America/Los_Angeles",
	}

	setLocalTimezone(t, time.UTC)

	expected := make(map[string]time.Time, len(inputs))
	for _, input := range inputs {
		parsed, err := Parse(input)
		if err != nil {
			t.Fatalf("Unable to parse date %q: %v", input, err)
		}
		expected[input] = parsed
	}

	for _, name := range []string{"America/Los_Angeles", "America/New_York", "Europe/Paris"} {
		t.Run(name, func(t *testing.T) {
			setLocalTimezone(t, loadLocation(t, name))

			for _, input := range inputs {
				parsed, err := Parse(input)
				if err != nil {
					t.Fatalf("Unable to parse date %q: %v", input, err)
				}

				if !parsed.Equal(expected[input]) {
					t.Errorf("Parse(%q) = %s, want %s as with time.Local set to UTC", input, parsed.Format(time.RFC3339), expected[input].Format(time.RFC3339))
				}
			}
		})
	}
}

func TestParseTimezoneOutOfRange(t *testing.T) {
	// Offsets go from -12:00 to +14:00; beyond that the date is converted to UTC.
	testCases := []struct {
		input      string
		want       string
		wantOffset int
	}{
		{"2023-05-29 00:00:00-13:00", "2023-05-29T13:00:00Z", 0},
		{"2023-05-29 00:00:00+15:00", "2023-05-28T09:00:00Z", 0},
		{"2023-05-29 00:00:00-12:00", "2023-05-29T12:00:00Z", -12 * 60 * 60},
		{"2023-05-29 00:00:00+14:00", "2023-05-28T10:00:00Z", 14 * 60 * 60},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			parsed, err := Parse(testCase.input)
			if err != nil {
				t.Fatalf("Unable to parse date %q: %v", testCase.input, err)
			}

			if want := mustParseRFC3339(t, testCase.want); !parsed.Equal(want) {
				t.Errorf("Parse(%q) = %s, want %s", testCase.input, parsed.Format(time.RFC3339), want.Format(time.RFC3339))
			}

			if _, offset := parsed.Zone(); offset != testCase.wantOffset {
				t.Errorf("The offset should be %d instead of %d", testCase.wantOffset, offset)
			}
		})
	}
}

func TestCheckTimezoneRange(t *testing.T) {
	testCases := []struct {
		name       string
		offset     int
		wantOffset int
	}{
		{"UTC", 0, 0},
		{"easternmost offset", 14 * 60 * 60, 14 * 60 * 60},
		{"beyond the easternmost offset", 14*60*60 + 60, 0},
		{"westernmost offset", -12 * 60 * 60, -12 * 60 * 60},
		{"beyond the westernmost offset", -12*60*60 - 60, 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			input := time.Date(2023, time.May, 29, 0, 0, 0, 0, time.FixedZone("", testCase.offset))
			got := checkTimezoneRange(input)

			if !got.Equal(input) {
				t.Errorf("The instant should be kept: got %s, want %s", got.Format(time.RFC3339), input.Format(time.RFC3339))
			}

			if _, offset := got.Zone(); offset != testCase.wantOffset {
				t.Errorf("The offset should be %d instead of %d", testCase.wantOffset, offset)
			}
		})
	}
}

func TestNormalizeRFC3339Separators(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{"2003-12-13t18:30:02z", "2003-12-13T18:30:02Z"},
		{"2003-12-13T18:30:02z", "2003-12-13T18:30:02Z"},
		{"2003-12-13t18:30:02Z", "2003-12-13T18:30:02Z"},
		{"2003-12-13t18:30:02+01:00", "2003-12-13T18:30:02+01:00"},
		{"2003-12-13T18:30:02Z", "2003-12-13T18:30:02Z"},
		{"2003-12-13t", "2003-12-13T"},
		{"2003-12-13", "2003-12-13"},
		{"", ""},
		{"2003/12/13t18:30:02z", "2003/12/13t18:30:02z"},
		{"2003-12-13 18:30:02z", "2003-12-13 18:30:02z"},
		{"Sat, 13 Dec 2003 18:30:02 z", "Sat, 13 Dec 2003 18:30:02 z"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			if got := normalizeRFC3339Separators(testCase.input); got != testCase.want {
				t.Errorf("normalizeRFC3339Separators(%q) = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}

func TestPadShortOffset(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{"Tue, Feb 16 2016 23:16:00 -700", "Tue, Feb 16 2016 23:16:00 -0700"},
		{"Sun, 28 Oct 2018 13:48:00 +530", "Sun, 28 Oct 2018 13:48:00 +0530"},
		{"Sun, 28 Oct 2018 13:48:00 -0700", "Sun, 28 Oct 2018 13:48:00 -0700"},
		{"Sun, 28 Oct 2018 13:48:00 -07", "Sun, 28 Oct 2018 13:48:00 -07"},
		{"Sun, 28 Oct 2018 13:48:00-700", "Sun, 28 Oct 2018 13:48:00-700"},
		{"Sun, 28 Oct 2018 13:48:00 -7a0", "Sun, 28 Oct 2018 13:48:00 -7a0"},
		{" -700", " -0700"},
		{"-700", "-700"},
		{"", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			if got := padShortOffset(testCase.input); got != testCase.want {
				t.Errorf("padShortOffset(%q) = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}
