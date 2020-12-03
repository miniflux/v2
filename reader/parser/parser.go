// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package parser // import "miniflux.app/reader/parser"

import (
	"strings"

	"miniflux.app/errors"
	"miniflux.app/model"
	"miniflux.app/reader/atom"
	"miniflux.app/reader/json"
	"miniflux.app/reader/rdf"
	"miniflux.app/reader/rss"
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
