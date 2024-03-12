// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"fmt"
	"io"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/xml"
)

// Parse returns a normalized feed struct from a RSS feed.
func Parse(baseURL string, data io.ReadSeeker) (*model.Feed, error) {
	feed := new(rssFeed)
	decoder := xml.NewXMLDecoder(data)
	decoder.DefaultSpace = "rss"
	if err := decoder.Decode(feed); err != nil {
		return nil, fmt.Errorf("rss: unable to parse feed: %w", err)
	}
	return feed.Transform(baseURL), nil
}
