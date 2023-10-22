// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rdf // import "miniflux.app/v2/internal/reader/rdf"

import (
	"fmt"
	"io"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/xml"
)

// Parse returns a normalized feed struct from a RDF feed.
func Parse(baseURL string, data io.Reader) (*model.Feed, error) {
	feed := new(rdfFeed)
	if err := xml.NewDecoder(data).Decode(feed); err != nil {
		return nil, fmt.Errorf("rdf: unable to parse feed: %w", err)
	}

	return feed.Transform(baseURL), nil
}
