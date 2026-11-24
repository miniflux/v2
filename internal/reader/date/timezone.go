// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package date // import "miniflux.app/v2/internal/reader/date"

import (
	"strings"
	"time"
)

// The offsets of real time zones go from UTC−12:00 to UTC+14:00, and PostgreSQL
// rejects dates outside of this range with "time zone displacement out of range".
// See https://en.wikipedia.org/wiki/List_of_UTC_offsets
const (
	minOffset = -12 * 60 * 60
	maxOffset = 14 * 60 * 60
)

// Dates are parsed in fixed locations, never in time.Local, so that the result
// does not depend on the time zone of the server. A location is nil when the
// time zone database is not available.
var (
	losAngelesLocation, _ = time.LoadLocation("America/Los_Angeles")
	newYorkLocation, _    = time.LoadLocation("America/New_York")
	brusselsLocation, _   = time.LoadLocation("Europe/Brussels")
)

type namedLocation struct {
	name     string
	location *time.Location
}

// timezoneAbbreviations lists the time zone abbreviations that are resolved in
// a location that uses them, when a date ends with one. Go gives any other
// abbreviation, except UTC and GMT, a zero offset.
var timezoneAbbreviations = [...]namedLocation{
	{"PST", losAngelesLocation},
	{"PDT", losAngelesLocation},
	{"EST", newYorkLocation},
	{"EDT", newYorkLocation},
}

// ianaTimezones lists the IANA time zone names found in some feed dates.
var ianaTimezones = [...]namedLocation{
	{"America/Los_Angeles", losAngelesLocation},
	{"Europe/Brussels", brusselsLocation},
}

// replaceIANATimezone replaces an IANA time zone name, which no layout accepts,
// with "UTC", which every layout with a time zone abbreviation accepts. It
// returns the location the parsed wall clock time belongs to, so that daylight
// saving time is taken into account, or nil when the value has no such name or
// the time zone database is not available.
func replaceIANATimezone(value string) (string, *time.Location) {
	for _, timezone := range ianaTimezones {
		if strings.Contains(value, timezone.name) {
			return strings.Replace(value, timezone.name, "UTC", 1), timezone.location
		}
	}

	return value, nil
}

// timezoneLocation returns the location to parse a date in. Go resolves a time
// zone abbreviation only with the zones of that location.
func timezoneLocation(value string) *time.Location {
	for _, timezone := range timezoneAbbreviations {
		if timezone.location != nil && strings.HasSuffix(value, timezone.name) {
			return timezone.location
		}
	}

	return time.UTC
}

// withWallClockIn returns the time with the same wall clock time in location.
func withWallClockIn(t time.Time, location *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), location)
}

// padShortOffset pads a numeric offset written with three digits at the end of
// a date, such as the "-700" of "Tue, Feb 16 2016 23:16:00 -700", to the four
// digits that Go accepts.
func padShortOffset(value string) string {
	offsetStart := len(value) - 4
	if offsetStart < 1 || value[offsetStart-1] != ' ' || (value[offsetStart] != '+' && value[offsetStart] != '-') {
		return value
	}

	for _, digit := range value[offsetStart+1:] {
		if digit < '0' || digit > '9' {
			return value
		}
	}

	return value[:offsetStart+1] + "0" + value[offsetStart+1:]
}

// checkTimezoneRange converts t to UTC when its offset is out of range.
func checkTimezoneRange(t time.Time) time.Time {
	if _, offset := t.Zone(); offset < minOffset || offset > maxOffset {
		return t.UTC()
	}

	return t
}
