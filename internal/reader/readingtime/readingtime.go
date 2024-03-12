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

	// Litterature on language detection says that around 100 signes is enough, we're safe here.
	truncationPoint := int(math.Min(float64(len(sanitizedContent)), 250))

	// We're only interested in identifying Japanse/Chinese/Korean
	options := whatlanggo.Options{
		Whitelist: map[whatlanggo.Lang]bool{
			whatlanggo.Jpn: true,
			whatlanggo.Cmn: true,
			whatlanggo.Kor: true,
		},
	}
	langInfo := whatlanggo.DetectWithOptions(sanitizedContent[:truncationPoint], options)

	if langInfo.IsReliable() {
		return int(math.Ceil(float64(utf8.RuneCountInString(sanitizedContent)) / float64(cjkReadingSpeed)))
	}
	nbOfWords := len(strings.Fields(sanitizedContent))
	return int(math.Ceil(float64(nbOfWords) / float64(defaultReadingSpeed)))
}
