// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package language // import "miniflux.app/v2/internal/reader/language"

import "strings"

// maxLength bounds accepted language tags. RFC 5646 recommends supporting
// tags of at least 35 characters; anything much longer is garbage.
const maxLength = 50

// Normalize cleans up a language tag declared by a feed so it is
// suitable for use as an HTML lang attribute. It trims surrounding
// whitespace, lower-cases the value, and replaces underscores with hyphens
// (e.g. "en_US" -> "en-us"). No strict BCP-47 validation is performed:
// many real feeds use loose values and silently dropping them yields worse
// downstream behaviour than passing them through.
//
// The value is feed-controlled and is persisted and rendered as-is, so
// anything outside the BCP-47 tag alphabet ([a-z0-9-]) or longer than
// maxLength is rejected: such a value carries no usable language
// information, and stripping bad characters could assemble a wrong tag.
func Normalize(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLength {
		return ""
	}

	// Lower-case ASCII-only, in the same pass as the charset check.
	// Unicode case folding (strings.ToLower) would map some non-ASCII
	// characters to ASCII (e.g. the Kelvin sign U+212A to "k"), turning
	// input the filter should reject into an apparently valid tag.
	b := []byte(s)
	for i, c := range b {
		switch {
		case c >= 'A' && c <= 'Z':
			b[i] = c + 'a' - 'A'
		case c == '_':
			b[i] = '-'
		case (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-':
		default:
			return ""
		}
	}
	return string(b)
}
