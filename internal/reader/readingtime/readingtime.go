// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package readtime provides a function to estimate the reading time of an article.
package readingtime

import (
	"math"
	"strings"
	"unicode/utf8"

	"miniflux.app/v2/internal/reader/sanitizer"

	"github.com/abadojack/whatlanggo"
)

// EstimateReadingTime returns the estimated reading time of an article in minute.
func EstimateReadingTime(content string, defaultReadingSpeed, cjkReadingSpeed int) int {
	sanitizedContent := sanitizer.StripTags(content)
	langInfo := whatlanggo.Detect(sanitizedContent)

	var timeToReadInt int
	if langInfo.IsReliable() && (langInfo.Lang == whatlanggo.Jpn || langInfo.Lang == whatlanggo.Cmn || langInfo.Lang == whatlanggo.Kor) {
		timeToReadInt = int(math.Ceil(float64(utf8.RuneCountInString(sanitizedContent)) / float64(cjkReadingSpeed)))
	} else {
		nbOfWords := len(strings.Fields(sanitizedContent))
		timeToReadInt = int(math.Ceil(float64(nbOfWords) / float64(defaultReadingSpeed)))
	}

	return timeToReadInt
}
