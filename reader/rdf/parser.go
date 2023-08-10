// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rdf // import "miniflux.app/v2/reader/rdf"

import (
	"io"

	"miniflux.app/v2/errors"
	"miniflux.app/v2/model"
	"miniflux.app/v2/reader/xml"
)

// Parse returns a normalized feed struct from a RDF feed.
func Parse(baseURL string, data io.Reader) (*model.Feed, *errors.LocalizedError) {
	feed := new(rdfFeed)
	decoder := xml.NewDecoder(data)
	err := decoder.Decode(feed)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse RDF feed: %q", err)
	}

	return feed.Transform(baseURL), nil
}
