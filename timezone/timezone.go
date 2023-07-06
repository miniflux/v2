// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package timezone // import "miniflux.app/timezone"

import (
	"time"
)

// Convert converts provided date time to actual timezone.
func Convert(tz string, t time.Time) time.Time {
	userTimezone := getLocation(tz)

	if t.Location().String() == "" {
		// In this case, the provided date is already converted to the user timezone by Postgres,
		// but the timezone information is not set in the time struct.
		// We cannot use time.In() because the date will be converted a second time.
		t = time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			t.Minute(),
			t.Second(),
			t.Nanosecond(),
			userTimezone,
		)
	} else if t.Location() != userTimezone {
		t = t.In(userTimezone)
	}

	return t
}

// Now returns the current time with the given timezone.
func Now(tz string) time.Time {
	return time.Now().In(getLocation(tz))
}

func getLocation(tz string) *time.Location {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.Local
	}
	return loc
}
