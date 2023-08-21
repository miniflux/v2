// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"io"

	"miniflux.app/v2/internal/errors"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/xml"
)

// Parse returns a normalized feed struct from a RSS feed.
func Parse(baseURL string, data io.Reader) (*model.Feed, *errors.LocalizedError) {
	feed := new(rssFeed)
	decoder := xml.NewDecoder(data)
	err := decoder.Decode(feed)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse RSS feed: %q", err)
	}

	return feed.Transform(baseURL), nil
}
