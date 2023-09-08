// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package dublincore // import "miniflux.app/v2/internal/reader/dublincore"

import (
	"strings"

	"miniflux.app/v2/internal/reader/sanitizer"
)

// DublinCoreFeedElement represents Dublin Core feed XML elements.
type DublinCoreFeedElement struct {
	DublinCoreCreator string `xml:"http://purl.org/dc/elements/1.1/ channel>creator"`
}

func (feed *DublinCoreFeedElement) GetSanitizedCreator() string {
	return strings.TrimSpace(sanitizer.StripTags(feed.DublinCoreCreator))
}

// DublinCoreItemElement represents Dublin Core entry XML elements.
type DublinCoreItemElement struct {
	DublinCoreDate    string `xml:"http://purl.org/dc/elements/1.1/ date"`
	DublinCoreCreator string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	DublinCoreContent string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
}

func (item *DublinCoreItemElement) GetSanitizedCreator() string {
	return strings.TrimSpace(sanitizer.StripTags(item.DublinCoreCreator))
}
