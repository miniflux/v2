// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rdf // import "miniflux.app/reader/rdf"

import (
	"encoding/xml"
	"io"

	"miniflux.app/errors"
	"miniflux.app/model"
	"miniflux.app/reader/encoding"
)

// Parse returns a normalized feed struct from a RDF feed.
func Parse(data io.Reader) (*model.Feed, *errors.LocalizedError) {
	feed := new(rdfFeed)
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = encoding.CharsetReader

	err := decoder.Decode(feed)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse RDF feed: %q", err)
	}

	return feed.Transform(), nil
}
