// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package atom

import (
	"encoding/xml"
	"fmt"
	"github.com/miniflux/miniflux2/model"
	"io"

	"golang.org/x/net/html/charset"
)

// Parse returns a normalized feed struct.
func Parse(data io.Reader) (*model.Feed, error) {
	atomFeed := new(AtomFeed)
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(atomFeed)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse Atom feed: %v\n", err)
	}

	return atomFeed.Transform(), nil
}
