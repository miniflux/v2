// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package date // import "miniflux.app/v2/internal/reader/date"

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Parse parses a date found in a feed: a Unix timestamp, or a date in one of the
// many formats listed in layouts, possibly with localized names.
func Parse(rawInput string) (time.Time, error) {
	input := strings.TrimSpace(rawInput)
	if input == "" {
		return time.Time{}, errors.New(`date parser: empty value`)
	}

	if timestamp, err := strconv.ParseInt(input, 10, 64); err == nil {
		return time.Unix(timestamp, 0), nil
	}

	value, wallClockLocation := replaceIANATimezone(input)
	value = padShortOffset(normalizeRFC3339Separators(replacer.Replace(value)))

	parsed, ok := parseLayouts(value)
	if !ok {
		return time.Time{}, fmt.Errorf(`date parser: failed to parse date "%s"`, input)
	}

	if wallClockLocation != nil {
		parsed = withWallClockIn(parsed, wallClockLocation)
	}

	return checkTimezoneRange(parsed), nil
}

// parseLayouts parses the value with the first layout that matches it.
func parseLayouts(value string) (time.Time, bool) {
	location := timezoneLocation(value)

	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

// normalizeRFC3339Separators upper-cases the "t" separator and the "z" time zone
// designator of RFC 3339 date-times. RFC 3339 section 5.6 allows them to be
// lower case, but time.Parse only accepts upper case.
func normalizeRFC3339Separators(value string) string {
	if len(value) < 11 || value[4] != '-' || value[7] != '-' || (value[10] != 't' && value[10] != 'T') {
		return value
	}

	value = value[:10] + "T" + value[11:]
	if strings.HasSuffix(value, "z") {
		value = value[:len(value)-1] + "Z"
	}

	return value
}
