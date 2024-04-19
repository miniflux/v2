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
	r.Seek(0, io.SeekStart)
	format, version := DetectFeedFormat(r)
	switch format {
	case FormatAtom:
		r.Seek(0, io.SeekStart)
		return atom.Parse(baseURL, r, version)
	case FormatRSS:
		r.Seek(0, io.SeekStart)
		return rss.Parse(baseURL, r)
	case FormatJSON:
		r.Seek(0, io.SeekStart)
		return json.Parse(baseURL, r)
	case FormatRDF:
		r.Seek(0, io.SeekStart)
		return rdf.Parse(baseURL, r)
	default:
		return nil, ErrFeedFormatNotDetected
	}
}
