// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer // import "miniflux.app/v2/internal/reader/sanitizer"

import (
	"errors"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// StripTags removes all HTML/XML tags from the input string.
// This function must *only* be used for cosmetic purposes, not to prevent code injections like XSS.
func StripTags(input string) string {
	dst := &strings.Builder{}
	src := strings.NewReader(input)

	err := stripIter(src, func(text string) bool {
		dst.WriteString(text)
		return true
	})
	if err != nil {
		return ""
	}

	return strings.TrimSpace(dst.String())
}

// stripIter iterates over the input [io.Reader] and calls the yield function for each [html.TextToken].
// Other kinds of [html.TokenType] are skipped.
func stripIter(src io.Reader, yield func(string) bool) error {
	tokenizer := html.NewTokenizer(src)

	for tokenizer.Next() != html.ErrorToken {
		token := tokenizer.Token()
		if token.Type != html.TextToken {
			continue
		}

		if !yield(token.Data) {
			break
		}
	}

	if err := tokenizer.Err(); !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
