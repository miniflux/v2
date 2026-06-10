// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// TruncateHTML returns cleaned up and shortened version of input.
//   - HTML tags are removed
//   - Consecutive whitespace characters replaced with single SPACE (0x20) character
//   - If input has more runes than limit, it's truncated
func TruncateHTML(input string, limit int) string {
	dst := &strings.Builder{}
	src := strings.NewReader(input)

	words := 0
	count := 0
	needspace := false

	err := stripIter(src, func(token string) bool {
		// Skip leading space.
		if words > 0 {
			// Add a space between tokens if there's one before HTML tag.
			r, _ := utf8.DecodeRuneInString(token)
			needspace = needspace || unicode.IsSpace(r)
		}

		for word := range strings.FieldsSeq(token) {
			if needspace {
				if count += 1; count > limit {
					return false
				}
			}

			// Compute how much of the word we can use later.
			wordlen := 0
			for wordlen < len(word) {
				if count += 1; count > limit {
					break
				}

				r, w := utf8.DecodeRuneInString(word[wordlen:])
				if r == utf8.RuneError {
					wordlen += 1
					continue
				}

				wordlen += w
			}

			// This is the only place where space being placed.
			// That way any sequence of space characters ends up as a singular SPACE (0x20) character.
			//
			// wordlen > 0 skips spaces if no printable characters left.
			if needspace && wordlen > 0 {
				dst.WriteByte(' ')
			}

			dst.WriteString(word[:wordlen])

			if count > limit {
				return false
			}

			needspace = true // To insert spaces in-between words in a token.
			words++
		}

		// Add a space between tokens if there's one after HTML tag.
		r, _ := utf8.DecodeLastRuneInString(token)
		needspace = unicode.IsSpace(r) && words > 0

		return true
	})
	if err != nil {
		return ""
	}

	if count > limit {
		dst.WriteRune('…')
	}

	return dst.String()
}
