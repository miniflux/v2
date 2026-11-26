// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package parser // import "miniflux.app/v2/internal/reader/parser"

import (
	"errors"
	"io"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/atom"
	"miniflux.app/v2/internal/reader/json"
	"miniflux.app/v2/internal/reader/rdf"
	"miniflux.app/v2/internal/reader/rss"
)

var ErrFeedFormatNotDetected = errors.New("parser: unable to detect feed format")

// ParseFeed analyzes the input data and returns a normalized feed object.
func ParseFeed(baseURL string, r io.ReadSeeker) (*model.Feed, error) {
	format, version := DetectFeedFormat(r)
	var feed *model.Feed
	var err error
	switch format {
	case FormatAtom:
		feed, err = atom.Parse(baseURL, r, version)
	case FormatRSS:
		feed, err = rss.Parse(baseURL, r)
	case FormatJSON:
		feed, err = json.Parse(baseURL, r)
	case FormatRDF:
		feed, err = rdf.Parse(baseURL, r)
	default:
		return nil, ErrFeedFormatNotDetected
	}
	if err != nil {
		return feed, err
	}

	// Record the publisher's version before rewriting, scraping, or sanitizing.
	for _, entry := range feed.Entries {
		entry.SourceFingerprint = entry.ComputeSourceFingerprint()
	}
	return feed, nil
}
