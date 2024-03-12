// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

var ErrInvalidDurationFormat = errors.New("rss: invalid duration format")

// normalizeDuration returns the duration tag value as a number of minutes
func normalizeDuration(rawDuration string) (int, error) {
	var sumSeconds int

	durationParts := strings.Split(rawDuration, ":")
	if len(durationParts) > 3 {
		return 0, ErrInvalidDurationFormat
	}

	for i, durationPart := range durationParts {
		durationPartValue, err := strconv.Atoi(durationPart)
		if err != nil {
			return 0, ErrInvalidDurationFormat
		}

		sumSeconds += int(math.Pow(60, float64(len(durationParts)-i-1))) * durationPartValue
	}

	return sumSeconds / 60, nil
}
