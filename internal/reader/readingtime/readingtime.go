// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package readingtime provides a function to estimate the reading time of an article.
package readingtime

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"miniflux.app/v2/internal/reader/sanitizer"
)

// EstimateReadingTime returns the estimated reading time of an article in minute.
func EstimateReadingTime(content string, defaultReadingSpeed, cjkReadingSpeed int) int {
	const truncationPoint = 100

	sanitizedContent := sanitizer.StripTags(content)

	if isCJK(sanitizedContent, truncationPoint) {
		return (utf8.RuneCountInString(sanitizedContent) + cjkReadingSpeed - 1) / cjkReadingSpeed
	}

	return (countWords(sanitizedContent) + defaultReadingSpeed - 1) / defaultReadingSpeed
}

func countWords(s string) int {
	n := 0
	for range strings.FieldsSeq(s) {
		n++
	}
	return n
}

func isCJK(text string, limit int) bool {
	var letters, totalCJK int
	for _, r := range text {
		// Numbers and control characters often used in CJK too.
		// Counting them makes detection less reliable.
		if !unicode.In(r, unicode.Letter) {
			continue
		}

		if letters++; letters == limit {
			break
		}

		if unicode.In(r, unicode.Han, unicode.Hangul, unicode.Hiragana, unicode.Katakana, unicode.Yi, unicode.Bopomofo) {
			totalCJK++
		}
	}

	// If at least half of the letters is CJK, odds are that the text is CJK.
	midpoint := letters / 2

	return totalCJK > midpoint
}
