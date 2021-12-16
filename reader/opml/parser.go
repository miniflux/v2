// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml // import "miniflux.app/reader/opml"

import (
	"encoding/xml"
	"io"

	"miniflux.app/errors"
	"miniflux.app/reader/encoding"
)

// Parse reads an OPML file and returns a SubcriptionList.
func Parse(data io.Reader) (SubcriptionList, *errors.LocalizedError) {
	opmlDocument := NewOPMLDocument()
	decoder := xml.NewDecoder(data)
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false
	decoder.CharsetReader = encoding.CharsetReader

	err := decoder.Decode(opmlDocument)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse OPML file: %q", err)
	}

	return opmlDocument.GetSubscriptionList(), nil
}
