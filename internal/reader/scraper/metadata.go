// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package scraper // import "miniflux.app/v2/internal/reader/scraper"

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractHeadMetadata walks the <head> section of the document and collects
// values from <meta> tags that are commonly used for previews (OpenGraph and
// Twitter Cards).
//
// The returned map keys are normalized to lowercase. OpenGraph properties are
// prefixed with "og:" and Twitter Card names with "twitter:", matching the
// attribute values used in the markup.
func ExtractHeadMetadata(document *goquery.Document) map[string]string {
	metadata := make(map[string]string)

	document.Find("head meta").Each(func(_ int, s *goquery.Selection) {
		content, hasContent := s.Attr("content")
		if !hasContent {
			return
		}
		content = strings.TrimSpace(content)
		if content == "" {
			return
		}

		// OpenGraph uses the "property" attribute; Twitter Cards historically
		// use the "name" attribute, but a number of sites use "property" for
		// Twitter tags too.
		for _, attr := range []string{"property", "name"} {
			key, hasKey := s.Attr(attr)
			if !hasKey {
				continue
			}
			key = strings.ToLower(strings.TrimSpace(key))
			if key == "" {
				continue
			}
			if !strings.HasPrefix(key, "og:") && !strings.HasPrefix(key, "twitter:") {
				continue
			}
			// First value wins so feed authors can rely on document order.
			if _, alreadySet := metadata[key]; !alreadySet {
				metadata[key] = content
			}
		}
	})

	return metadata
}
