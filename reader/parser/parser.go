// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package parser // import "miniflux.app/v2/reader/parser"

import (
	"strings"

	"miniflux.app/v2/errors"
	"miniflux.app/v2/model"
	"miniflux.app/v2/reader/atom"
	"miniflux.app/v2/reader/json"
	"miniflux.app/v2/reader/rdf"
	"miniflux.app/v2/reader/rss"
)

// ParseFeed analyzes the input data and returns a normalized feed object.
func ParseFeed(baseURL, data string) (*model.Feed, *errors.LocalizedError) {
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
		return nil, errors.NewLocalizedError("Unsupported feed format")
	}
}
