// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package date // import "miniflux.app/reader/date"

import (
	"errors"
	"fmt"
	"math"
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
	"Mon, 02 Jan 2006 15:04:05 MST -07:00",
	"Mon, January 2, 2006, 3:04 PM MST",
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
	"Mon, 02/01/2006",
	"Monday, 2. January 2006 - 15:04",
	"Monday 02 January 2006",
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
	"Mon Jan 02 2006 15:04:05 MST",
	"Mon, 01 02 2006 15:04:05 -0700",
	"Mon, 2th Jan 2006 15:05:05 MST",
	"Jan. 2, 2006, 3:04 a.m.",
	"fri, 02 jan 2006 15:04:05 -0700",
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
	"2006-01-02T15:04:05-07:00Z",
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
	"Jan. 2006",
	"Jan. 2, 2006, 03:04 p.m.",
	"2006-01-02 15:04:05 -07:00",
	"2 January, 2006",
	"2 Jan 2006 MST",
	"Mon, January 2, 2006 at 03:04 PM MST",
	"Jan 2, 2006 15:04 MST",
	"01/02/2006 3:04 pm MST",
	"Mon, 2th Jan 2006 15:04:05 MST",
	"Mon, 2rd Jan 2006 15:04:05 MST",
	"Mon, 2nd Jan 2006 15:04:05 MST",
	"Mon, 2st Jan 2006 15:04:05 MST",
}

var invalidTimezoneReplacer = strings.NewReplacer(
	"Europe/Brussels", "CET",
	"GMT+0000 (Coordinated Universal Time)", "GMT",
	"GMT-", "GMT -",
)

var invalidLocalizedDateReplacer = strings.NewReplacer(
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
	"lun.", "Mon",
	"mar.", "Tue",
	"mer.", "Wed",
	"jeu.", "Thu",
	"ven.", "Fri",
	"sam.", "Sat",
	"dim.", "Sun",
	"Lundi,", "Monday,",
	"Mardi,", "Tuesday,",
	"Mercredi,", "Wednesday,",
	"Jeudi,", "Thursday,",
	"Vendredi,", "Friday,",
	"Samedi,", "Saturday,",
	"Dimanche,", "Sunday,",
	"jan.", "January ",
	"feb.", "February ",
	"mars.", "March ",
	"avril.", "April ",
	"mai.", "May ",
	"juin.", "June ",
	"juil.", "July",
	"août.", "August",
	"sept.", "September",
	"oct.", "October",
	"nov.", "November",
	"dec.", "December",
	"déc.", "December",
	"janvier ", "January ",
	"février ", "February ",
	"mars ", "March ",
	"avril ", "April ",
	"mai ", "May ",
	"juin ", "June ",
	"juillet ", "July",
	"août ", "August",
	"septembre ", "September",
	"octobre ", "October",
	"november ", "November",
	"décembre ", "December",
	"Janvier", "January",
	"Février", "February",
	"Mars", "March",
	"Avril", "April",
	"Mai", "May",
	"Juin", "June",
	"Juillet", "July",
	"Août", "August",
	"Septembre", "September",
	"Octobre", "October",
	"Novembre", "November",
	"Décembre", "December",
	"avr ", "Apr ",
	"mai ", "May ",
	"jui ", "Jun ",
	"juin ", "June ",
	"Thurs,", "Thu,",
	"Thur,", "Thu,",
)

// Parse parses a given date string using a large
// list of commonly found feed date formats.
func Parse(rawInput string) (t time.Time, err error) {
	timestamp, err := strconv.ParseInt(rawInput, 10, 64)
	if err == nil {
		return time.Unix(timestamp, 0), nil
	}

	processedInput := invalidLocalizedDateReplacer.Replace(rawInput)
	processedInput = invalidTimezoneReplacer.Replace(processedInput)
	processedInput = strings.TrimSpace(processedInput)
	if processedInput == "" {
		return t, errors.New(`date parser: empty value`)
	}

	for _, layout := range dateFormats {
		switch layout {
		case time.RFC822, time.RFC850, time.RFC1123:
			if t, err = parseLocalTimeDates(layout, processedInput); err == nil {
				t = checkTimezoneRange(t)
				return
			}
		}

		if t, err = time.Parse(layout, processedInput); err == nil {
			t = checkTimezoneRange(t)
			return
		}
	}

	err = fmt.Errorf(`date parser: failed to parse date "%s"`, rawInput)
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
	if strings.HasSuffix(ds, "PST") || strings.HasSuffix(ds, "PDT") {
		loc, _ = time.LoadLocation("America/Los_Angeles")
	}

	if strings.HasSuffix(ds, "EST") || strings.HasSuffix(ds, "EDT") {
		loc, _ = time.LoadLocation("America/New_York")
	}

	return time.ParseInLocation(layout, ds, loc)
}

// https://en.wikipedia.org/wiki/List_of_UTC_offsets
// Offset range: westernmost (−12:00) to the easternmost (+14:00)
// Avoid "pq: time zone displacement out of range" errors
func checkTimezoneRange(t time.Time) time.Time {
	_, offset := t.Zone()
	if math.Abs(float64(offset)) > 14*60*60 {
		t = t.UTC()
	}
	return t
}
