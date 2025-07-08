// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

// parseISO8601Duration parses a subset of ISO8601 durations, mainly for youtube video.
func parseISO8601Duration(duration string) (time.Duration, error) {
	after, ok := strings.CutPrefix(duration, "PT")
	if !ok {
		return 0, errors.New("the period doesn't start with PT")
	}

	var d time.Duration
	num := ""

	for _, char := range after {
		var val float64
		var err error

		switch char {
		case 'Y', 'W', 'D':
			return 0, fmt.Errorf("the '%c' specifier isn't supported", char)
		case 'H':
			if val, err = strconv.ParseFloat(num, 64); err != nil {
				return 0, err
			}
			d += time.Duration(val) * time.Hour
			num = ""
		case 'M':
			if val, err = strconv.ParseFloat(num, 64); err != nil {
				return 0, err
			}
			d += time.Duration(val) * time.Minute
			num = ""
		case 'S':
			if val, err = strconv.ParseFloat(num, 64); err != nil {
				return 0, err
			}
			d += time.Duration(val) * time.Second
			num = ""
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
			num += string(char)
			continue
		default:
			return 0, errors.New("invalid character in the period")
		}
	}
	return d, nil
}

func minifyContent(content string) string {
	m := minify.New()

	// Options required to avoid breaking the HTML content.
	m.Add("text/html", &html.Minifier{
		KeepEndTags:         true,
		KeepQuotes:          true,
		KeepComments:        false,
		KeepSpecialComments: false,
		KeepDefaultAttrVals: false,
	})

	if minifiedHTML, err := m.String("text/html", content); err == nil {
		content = minifiedHTML
	}

	return content
}
