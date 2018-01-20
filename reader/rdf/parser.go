// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rdf

import (
	"encoding/xml"
	"io"

	"github.com/miniflux/miniflux/errors"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/reader/encoding"
)

// Parse returns a normalized feed struct from a RDF feed.
func Parse(data io.Reader) (*model.Feed, error) {
	feed := new(rdfFeed)
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = encoding.CharsetReader

	err := decoder.Decode(feed)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse RDF feed: %v.", err)
	}

	return feed.Transform(), nil
}
