// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package language // import "miniflux.app/v2/internal/reader/language"

import "strings"

// Normalize cleans up a language tag declared by a feed so it is
// suitable for use as an HTML lang attribute. It trims surrounding
// whitespace, lower-cases the value, and replaces underscores with hyphens
// (e.g. "en_US" -> "en-us"). No strict BCP-47 validation is performed:
// many real feeds use loose values and silently dropping them yields worse
// downstream behaviour than passing them through.
func Normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.ReplaceAll(s, "_", "-")
}
