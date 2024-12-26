// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package readingtime provides a function to estimate the reading time of an article.
package readingtime

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"miniflux.app/v2/internal/reader/sanitizer"
)

// EstimateReadingTime returns the estimated reading time of an article in minute.
func EstimateReadingTime(content string, defaultReadingSpeed, cjkReadingSpeed int) int {
	sanitizedContent := sanitizer.StripTags(content)
	truncationPoint := min(len(sanitizedContent), 50)

	if isCJK(sanitizedContent[:truncationPoint]) {
		return int(math.Ceil(float64(utf8.RuneCountInString(sanitizedContent)) / float64(cjkReadingSpeed)))
	}
	return int(math.Ceil(float64(len(strings.Fields(sanitizedContent))) / float64(defaultReadingSpeed)))
}

func isCJK(text string) bool {
	totalCJK := 0

	for _, r := range text[:min(len(text), 50)] {
		if unicode.Is(unicode.Han, r) ||
			unicode.Is(unicode.Hangul, r) ||
			unicode.Is(unicode.Hiragana, r) ||
			unicode.Is(unicode.Katakana, r) ||
			unicode.Is(unicode.Yi, r) ||
			unicode.Is(unicode.Bopomofo, r) {
			totalCJK++
		}
	}

	// if at least 50% of the text is CJK, odds are that the text is in CJK.
	return totalCJK > len(text)/50
}
