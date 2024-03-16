// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

// FeedBurnerItemElement represents FeedBurner XML elements.
type FeedBurnerItemElement struct {
	FeedBurnerLink          string `xml:"http://rssnamespace.org/feedburner/ext/1.0 origLink"`
	FeedBurnerEnclosureLink string `xml:"http://rssnamespace.org/feedburner/ext/1.0 origEnclosureLink"`
}
