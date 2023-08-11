// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer

import "strings"

func TruncateHTML(input string, max int) string {
	text := StripTags(input)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "  ", " ")
	text = strings.TrimSpace(text)

	// Convert to runes to be safe with unicode
	runes := []rune(text)
	if len(runes) > max {
		return strings.TrimSpace(string(runes[:max])) + "…"
	}

	return text
}
