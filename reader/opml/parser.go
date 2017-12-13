// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml

import (
	"encoding/xml"
	"io"

	"github.com/miniflux/miniflux/errors"
	"golang.org/x/net/html/charset"
)

// Parse reads an OPML file and returns a SubcriptionList.
func Parse(data io.Reader) (SubcriptionList, error) {
	feeds := new(opml)
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(feeds)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse OPML file: %v.", err)
	}

	return feeds.Transform(), nil
}
