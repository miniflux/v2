// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package parser // import "miniflux.app/v2/internal/reader/parser"

import (
	"errors"
	"strings"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/atom"
	"miniflux.app/v2/internal/reader/json"
	"miniflux.app/v2/internal/reader/rdf"
	"miniflux.app/v2/internal/reader/rss"
)

var ErrFeedFormatNotDetected = errors.New("parser: unable to detect feed format")

// ParseFeed analyzes the input data and returns a normalized feed object.
func ParseFeed(baseURL, data string) (*model.Feed, error) {
	switch DetectFeedFormat(data) {
	case FormatAtom:
		return atom.Parse(baseURL, strings.NewReader(data))
	case FormatRSS:
		return rss.Parse(baseURL, strings.NewReader(data))
	case FormatJSON:
		return json.Parse(baseURL, strings.NewReader(data))
	case FormatRDF:
		return rdf.Parse(baseURL, strings.NewReader(data))
	default:
		return nil, ErrFeedFormatNotDetected
	}
}
