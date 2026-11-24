// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package date // import "miniflux.app/v2/internal/reader/date"

import "time"

// layouts lists the date formats found in feeds, after the replacer has
// translated localized names. Parse tries them in order and keeps the first one
// that matches.
//
// The order only matters between layouts that can match the same input. When a
// day-first and a month-first layout differ only in that order, the day-first
// one comes first, so "02/01/2006" reads as 2 January.
//
// When adding a layout, keep in mind how time.Parse reads it:
//   - "2", "1", "3", "4" and "5" accept one or two digits, so a "02" copy of an
//     earlier "2" layout is never reached.
//   - A space matches any run of spaces.
//   - Day and month names match in any case.
//   - Anything that is not a layout element must appear as is: "14", "6" and
//     "-700" are literal text, not an hour, a year or an offset.
//
// Add a test case to validDates for each new layout.
var layouts = [...]string{
	// Go standard layouts.
	time.RFC822, // RSS
	time.RFC850,
	time.RFC1123,
	time.RFC822Z, // RSS
	time.RFC3339, // Atom
	time.UnixDate,
	time.RubyDate,
	time.RFC1123Z,
	time.ANSIC,

	// Abbreviated weekday, then month name and day.
	"Mon, January 2, 2006, 3:04 PM MST",
	"Mon, January 2 2006 15:04:05 -0700",
	"Mon, January 02, 2006, 15:04:05 MST",
	"Mon, January 02, 2006 15:04:05 MST",
	"Mon, Jan 2, 2006 15:04 MST",
	"Mon, Jan 2 2006 15:04 MST",
	"Mon, Jan 2 2006 15:04:05 MST",
	"Mon, Jan 2, 2006 15:04:05 MST",
	"Mon, Jan 2 2006 15:04:05 -0700",
	"Mon Jan 2 15:04 2006",
	"Mon Jan 2 15:04:05 2006 MST",
	"Mon Jan 02, 2006 3:04 pm",
	"Mon, Jan 02,2006 15:04:05 MST",
	"Mon Jan 02 2006 15:04:05 -0700",
	"Mon Jan 02 2006 15:04:05 MST",
	"Mon, January 2, 2006 at 03:04 PM MST",
	"Mon, Jan 02 2006 03:04:05 PM",

	// Abbreviated weekday, then day and month name.
	"Mon, 02 Jan 2006 15:04:05 MST -07:00",
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
	"Mon 2 Jan 2006 15:04:05 MST",
	"Mon,2 Jan 2006 15:04:05 MST",
	"Mon, 2 Jan 2006 15:04:05 -0700 MST",
	"Mon, 2 Jan 2006 15:04:05-0700",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05",
	"Mon, 2 Jan 2006 15:04",
	"Mon, 02 Jan 2006, 15:04",
	"Mon, 2 Jan 2006, 15:04",
	"Mon,2 Jan 2006",
	"Mon, 2 Jan 2006",
	"Mon, 2 Jan 06 15:04:05 MST",
	"Mon, 2 Jan 06 15:04:05 -0700",
	"Mon,02 January 2006 15:04:05 MST",
	"Mon, 02 Jan 2006 3:04:05 PM MST",
	"Mon, 02 Jan 2006 15 -0700",
	"Mon,02 Jan 2006 15:04 MST",
	"Mon, 02 Jan 2006 15:04:05 Z",
	"Mon, 02 Jan 2006 15:04:05 MST-07:00",
	"Mon, 02 Jan 2006 15:04:05 MST -0700",
	"Mon, 02 Jan 2006, 15:04:05 MST",
	"Mon , 02 Jan 2006 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04:05 GMT-0700",
	"Mon,02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 -07:00",
	"Mon, 02 Jan 2006 15:04:05 --0700",
	"Mon 02 Jan 2006 15:04:05 -0700",
	"Mon 02 Jan 2006, 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04:05 -07",
	"Mon, 02 Jan 2006 15:04:05 00",
	"Mon, 02 Jan 2006 3:04 PM MST",
	"Mon, 2th Jan 2006 15:04:05 MST",
	"Mon, 2rd Jan 2006 15:04:05 MST",
	"Mon, 2nd Jan 2006 15:04:05 MST",
	"Mon, 2st Jan 2006 15:04:05 MST",

	// Abbreviated weekday, then numeric date.
	"Mon, 02/01/2006",
	"Mon, 01/02/2006 - 15:04",
	"Mon, 2006-01-02 15:04",
	"Mon, 02 01 2006 15:04:05 -0700",

	// Full weekday.
	"Monday, 2. January 2006 - 15:04",
	"Monday 02 January 2006",
	"Monday, January 2, 2006 15:04:05 MST",
	"Monday, January 2, 2006 03:04 PM",
	"Monday, January 2, 2006",
	"Monday, 2 January 2006 15:04:05 MST",
	"Monday, 2 January 2006 15:04:05 -0700",
	"Monday, 2 Jan 2006 15:04:05 MST",
	"Monday, 2 Jan 2006 15:04:05 -0700",
	"Monday, 02 January 2006 15:04:05",
	"Monday, January 02, 2006 - 3:04pm",
	"Monday, January 2, 2006 - 3:04pm",
	"Monday, January 2, 2006 - 15:04",

	// Month name first.
	"Jan. 2, 2006, 3:04 PM",
	"January 02 2006 03:04:05 PM",
	"January 2, 2006 3:04 PM",
	"January 2, 2006, 3:04 PM",
	"January 2, 2006 15:04:05 MST",
	"January 2, 2006 15:04:05",
	"January 2, 2006",
	"January 02, 2006 15:04",
	"Jan 2, 2006 3:04:05 PM MST",
	"Jan 2, 2006 3:04:05 PM",
	"Jan 2, 2006 15:04:05 MST",
	"Jan 2, 2006",
	"Jan 02 2006 03:04:05PM",
	"Jan. 2006",
	"Jan 2, 2006 15:04 MST",

	// Day, then month name.
	"2 January 2006 15:04:05 MST",
	"2 January 2006 15:04:05 -0700",
	"2 January 2006",
	"2 Jan 2006 15:04:05 Z",
	"2 Jan 2006 15:04:05 MST",
	"2 Jan 2006 15:04:05 -0700",
	"2 Jan 2006",
	"2 Jan 2006 15:04 MST",
	"2006 January 02",
	"02 Monday, Jan 2006 15:04",
	"02 Jan 2006 15:04:05 UT",
	"02 Jan 2006 15:04:05",
	"2 January, 2006",
	"2 Jan 2006 MST",

	// Year first.
	"2006-1-2T15:04:05Z",
	"2006-1-2 15:04:05",
	"2006-1-2",
	"2006-01-02T15:04:05-07:00Z",
	"2006-01-02T15:04Z",
	"2006-01-02T15:04-07:00",
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
	"2006-01-02 15:04:05 -07:00",

	// Numeric dates, and time first.
	"2.1.2006 15:04:05",
	"2/1/2006",
	"2-1-2006",
	"15:04 02.01.2006 -0700",
	"1/2/2006 3:04 PM MST",
	"1/2/2006 3:04:05 PM MST",
	"1/2/2006 3:04:05 PM",
	"1/2/2006 15:04:05 MST",
	"1/2/2006",
	"06/1/2 15:04",
	"06-1-2 15:04",
	"02/01/2006 15:04 MST",
	"02-01-2006 15:04:05 MST",
	"02/01/2006 15:04:05",
	"02.01.2006 15:04",
	"02/01/2006 - 15:04",
	"02.01.2006 -0700",
	"01/02/2006 3:04 PM",
	"01/02/2006 - 15:04",
	"01-02-2006",
	"01/02/2006 3:04 pm MST",
	"01/02/06 15:04:05",
	"02.01.06",
}
