// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package atom

import (
	"encoding/xml"
	"io"

	"github.com/miniflux/miniflux2/errors"
	"github.com/miniflux/miniflux2/model"

	"golang.org/x/net/html/charset"
)

// Parse returns a normalized feed struct from a Atom feed.
func Parse(data io.Reader) (*model.Feed, error) {
	atomFeed := new(atomFeed)
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(atomFeed)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse Atom feed: %v.", err)
	}

	return atomFeed.Transform(), nil
}
