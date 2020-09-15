// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package date // import "miniflux.app/reader/date"

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DateFormats taken from github.com/mjibson/goread
var dateFormats = []string{
	time.RFC822,  // RSS
	time.RFC822Z, // RSS
	time.RFC3339, // Atom
	time.UnixDate,
	time.RubyDate,
	time.RFC850,
	time.RFC1123Z,
	time.RFC1123,
	time.ANSIC,
	"Mon, January 2 2006 15:04:05 -0700",
	"Mon, January 02, 2006, 15:04:05 MST",
	"Mon, January 02, 2006 15:04:05 MST",
	"Mon, Jan 2, 2006 15:04 MST",
	"Mon, Jan 2 2006 15:04 MST",
	"Mon, Jan 2 2006 15:04:05 MST",
	"Mon, Jan 2, 2006 15:04:05 MST",
	"Mon, Jan 2 2006 15:04:05 -700",
	"Mon, Jan 2 2006 15:04:05 -0700",
	"Mon Jan 2 15:04 2006",
	"Mon Jan 2 15:04:05 2006 MST",
	"Mon Jan 02, 2006 3:04 pm",
	"Mon, Jan 02,2006 15:04:05 MST",
	"Mon Jan 02 2006 15:04:05 -0700",
	"Monday, January 2, 2006 15:04:05 MST",
	"Monday, January 2, 2006 03:04 PM",
	"Monday, January 2, 2006",
	"Monday, January 02, 2006",
	"Monday, 2 January 2006 15:04:05 MST",
	"Monday, 2 January 2006 15:04:05 -0700",
	"Monday, 2 Jan 2006 15:04:05 MST",
	"Monday, 2 Jan 2006 15:04:05 -0700",
	"Monday, 02 January 2006 15:04:05 MST",
	"Monday, 02 January 2006 15:04:05 -0700",
	"Monday, 02 January 2006 15:04:05",
	"Monday, January 02, 2006 - 3:04pm",
	"Monday, January 2, 2006 - 3:04pm",
	"Mon, 01/02/2006 - 15:04",
	"Mon, 2 January 2006 15:04 MST",
	"Mon, 2 January 2006, 15:04 -0700",
	"Mon, 2 January 2006, 15:04:05 MST",
	"Mon, 2 January 2006 15:04:05 MST",
	"Mon, 2 January 2006 15:04:05 -0700",
	"Mon, 2 January 2006",
	"Mon, 2 Jan 2006 3:04:05 PM -0700",
	"Mon, 2 Jan 2006 15:4:5 MST",
	"Mon, 2 Jan 2006 15:4:5 -0700 GMT",
	"Mon, 2, Jan 2006 15:4",
	"Mon, 2 Jan 2006 15:04 MST",
	"Mon, 2 Jan 2006, 15:04 -0700",
	"Mon, 2 Jan 2006 15:04 -0700",
	"Mon, 2 Jan 2006 15:04:05 UT",
	"Mon, 2 Jan 2006 15:04:05MST",
	"Mon, 2 Jan 2006 15:04:05 MST",
	"Mon 2 Jan 2006 15:04:05 MST",
	"mon,2 Jan 2006 15:04:05 MST",
	"Mon, 2 Jan 2006 15:04:05 -0700 MST",
	"Mon, 2 Jan 2006 15:04:05-0700",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05",
	"Mon, 2 Jan 2006 15:04",
	"Mon, 02 Jan 2006, 15:04",
	"Mon, 2 Jan 2006, 15:04",
	"Mon,2 Jan 2006",
	"Mon, 2 Jan 2006",
	"Mon, 2 Jan 15:04:05 MST",
	"Mon, 2 Jan 06 15:04:05 MST",
	"Mon, 2 Jan 06 15:04:05 -0700",
	"Mon, 2006-01-02 15:04",
	"Mon,02 January 2006 14:04:05 MST",
	"Mon, 02 January 2006",
	"Mon, 02 Jan 2006 3:04:05 PM MST",
	"Mon, 02 Jan 2006 15 -0700",
	"Mon,02 Jan 2006 15:04 MST",
	"Mon, 02 Jan 2006 15:04 MST",
	"Mon, 02 Jan 2006 15:04 -0700",
	"Mon, 02 Jan 2006 15:04:05 Z",
	"Mon, 02 Jan 2006 15:04:05 UT",
	"Mon, 02 Jan 2006 15:04:05 MST-07:00",
	"Mon, 02 Jan 2006 15:04:05 MST -0700",
	"Mon, 02 Jan 2006, 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04:05MST",
	"Mon, 02 Jan 2006 15:04:05 MST",
	"Mon , 02 Jan 2006 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04:05 GMT-0700",
	"Mon,02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 -07:00",
	"Mon, 02 Jan 2006 15:04:05 --0700",
	"Mon 02 Jan 2006 15:04:05 -0700",
	"Mon 02 Jan 2006, 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04:05 -07",
	"Mon, 02 Jan 2006 15:04:05 00",
	"Mon, 02 Jan 2006 15:04:05",
	"Mon, 02 Jan 2006",
	"Mon, 02 Jan 06 15:04:05 MST",
	"Mon, 02 Jan 2006 3:04 PM MST",
	"January 02 2006 03:04:05 PM",
	"January 2, 2006 3:04 PM",
	"January 2, 2006, 3:04 p.m.",
	"January 2, 2006 15:04:05 MST",
	"January 2, 2006 15:04:05",
	"January 2, 2006 03:04 PM",
	"January 2, 2006",
	"January 02, 2006 15:04:05 MST",
	"January 02, 2006 15:04",
	"January 02, 2006 03:04 PM",
	"January 02, 2006",
	"Jan 2, 2006 3:04:05 PM MST",
	"Jan 2, 2006 3:04:05 PM",
	"Jan 2, 2006 15:04:05 MST",
	"Jan 2, 2006",
	"Jan 02 2006 03:04:05PM",
	"Jan 02, 2006",
	"6/1/2 15:04",
	"6-1-2 15:04",
	"2 January 2006 15:04:05 MST",
	"2 January 2006 15:04:05 -0700",
	"2 January 2006",
	"2 Jan 2006 15:04:05 Z",
	"2 Jan 2006 15:04:05 MST",
	"2 Jan 2006 15:04:05 -0700",
	"2 Jan 2006",
	"2 Jan 2006 15:04 MST",
	"2.1.2006 15:04:05",
	"2/1/2006",
	"2-1-2006",
	"2006 January 02",
	"2006-1-2T15:04:05Z",
	"2006-1-2 15:04:05",
	"2006-1-2",
	"2006-1-02T15:04:05Z",
	"2006-01-02T15:04Z",
	"2006-01-02T15:04-07:00",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05-07:00:00",
	"2006-01-02T15:04:05:-0700",
	"2006-01-02T15:04:05-0700",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02T15:04:05 -0700",
	"2006-01-02T15:04:05:00",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 at 15:04:05",
	"2006-01-02 15:04:05Z",
	"2006-01-02 15:04:05 MST",
	"2006-01-02 15:04:05-0700",
	"2006-01-02 15:04:05-07:00",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04",
	"2006-01-02 00:00:00.0 15:04:05.0 -0700",
	"2006/01/02",
	"2006-01-02",
	"15:04 02.01.2006 -0700",
	"1/2/2006 3:04 PM MST",
	"1/2/2006 3:04:05 PM MST",
	"1/2/2006 3:04:05 PM",
	"1/2/2006 15:04:05 MST",
	"1/2/2006",
	"06/1/2 15:04",
	"06-1-2 15:04",
	"02 Monday, Jan 2006 15:04",
	"02 Jan 2006 15:04 MST",
	"02 Jan 2006 15:04:05 UT",
	"02 Jan 2006 15:04:05 MST",
	"02 Jan 2006 15:04:05 -0700",
	"02 Jan 2006 15:04:05",
	"02 Jan 2006",
	"02/01/2006 15:04 MST",
	"02-01-2006 15:04:05 MST",
	"02.01.2006 15:04:05",
	"02/01/2006 15:04:05",
	"02.01.2006 15:04",
	"02/01/2006 - 15:04",
	"02.01.2006 -0700",
	"02/01/2006",
	"02-01-2006",
	"01/02/2006 3:04 PM",
	"01/02/2006 15:04:05 MST",
	"01/02/2006 - 15:04",
	"01/02/2006",
	"01-02-2006",
}

// Parse parses a given date string using a large
// list of commonly found feed date formats.
func Parse(ds string) (t time.Time, err error) {
	timestamp, err := strconv.ParseInt(ds, 10, 64)
	if err == nil {
		return time.Unix(timestamp, 0), nil
	}

	ds = replaceNonEnglishWords(ds)
	d := strings.TrimSpace(ds)
	if d == "" {
		return t, errors.New("date parser: empty value")
	}

	for _, layout := range dateFormats {
		switch layout {
		case time.RFC822, time.RFC850, time.RFC1123:
			if t, err = parseLocalTimeDates(layout, d); err == nil {
				return
			}
		}

		if t, err = time.Parse(layout, d); err == nil {
			return
		}
	}

	lastSpace := strings.LastIndex(ds, " ")
	if lastSpace > 0 {
		return Parse(ds[0:lastSpace])
	}

	err = fmt.Errorf(`date parser: failed to parse date "%s"`, ds)
	return
}

// According to Golang documentation:
//
// RFC822, RFC850, and RFC1123 formats should be applied only to local times.
// Applying them to UTC times will use "UTC" as the time zone abbreviation,
// while strictly speaking those RFCs require the use of "GMT" in that case.
func parseLocalTimeDates(layout, ds string) (t time.Time, err error) {
	loc := time.UTC

	// Workaround for dates that don't use GMT.
	if strings.HasSuffix(ds, "PST") {
		loc, _ = time.LoadLocation("America/Los_Angeles")
	}

	return time.ParseInLocation(layout, ds, loc)
}

// Replace German and French dates to English.
func replaceNonEnglishWords(ds string) string {
	r := strings.NewReplacer(
		"Mo,", "Mon,",
		"Di,", "Tue,",
		"Mi,", "Wed,",
		"Do,", "Thu,",
		"Fr,", "Fri,",
		"Sa,", "Sat,",
		"So,", "Sun,",
		"Mär ", "Mar ",
		"Mai ", "May ",
		"Okt ", "Oct ",
		"Dez ", "Dec ",
		"lun,", "Mon,",
		"mar,", "Tue,",
		"mer,", "Wed,",
		"jeu,", "Thu,",
		"ven,", "Fri,",
		"sam,", "Sat,",
		"dim,", "Sun,",
		"avr ", "Apr ",
		"mai ", "May ",
		"jui ", "Jun ",
	)

	return r.Replace(ds)
}
