// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package atom

import (
	"encoding/xml"
	"io"

	"github.com/miniflux/miniflux/errors"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/reader/encoding"
)

// Parse returns a normalized feed struct from a Atom feed.
func Parse(data io.Reader) (*model.Feed, *errors.LocalizedError) {
	atomFeed := new(atomFeed)
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = encoding.CharsetReader

	err := decoder.Decode(atomFeed)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse Atom feed: %q", err)
	}

	return atomFeed.Transform(), nil
}
