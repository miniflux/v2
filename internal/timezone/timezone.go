// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package timezone // import "miniflux.app/v2/internal/timezone"

import (
	"sync"
	"time"
)

var (
	tzCache = sync.Map{} // Cache for time locations to avoid loading them multiple times.
)

// Convert converts provided date time to actual timezone.
func Convert(tz string, t time.Time) time.Time {
	userTimezone := getLocation(tz)

	if t.Location().String() == "" {
		if t.Before(time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)) {
			return time.Date(0, time.January, 1, 0, 0, 0, 0, userTimezone)
		}

		// In this case, the provided date is already converted to the user timezone by Postgres,
		// but the timezone information is not set in the time struct.
		// We cannot use time.In() because the date will be converted a second time.
		return time.Date(
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
		return t.In(userTimezone)
	}

	return t
}

// Now returns the current time with the given timezone.
func Now(tz string) time.Time {
	return time.Now().In(getLocation(tz))
}

func getLocation(tz string) *time.Location {
	if loc, ok := tzCache.Load(tz); ok {
		return loc.(*time.Location)
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.Local
	}

	tzCache.Store(tz, loc)
	return loc
}
