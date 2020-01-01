// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package atom // import "miniflux.app/reader/atom"

import (
	"bytes"
	"encoding/xml"
	"io"

	"miniflux.app/errors"
	"miniflux.app/model"
	xml_decoder "miniflux.app/reader/xml"
)

type atomFeed interface {
	Transform() *model.Feed
}

// Parse returns a normalized feed struct from a Atom feed.
func Parse(r io.Reader) (*model.Feed, *errors.LocalizedError) {
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)

	var rawFeed atomFeed
	if getAtomFeedVersion(tee) == "0.3" {
		rawFeed = new(atom03Feed)
	} else {
		rawFeed = new(atom10Feed)
	}

	decoder := xml_decoder.NewDecoder(&buf)
	err := decoder.Decode(rawFeed)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse Atom feed: %q", err)
	}

	return rawFeed.Transform(), nil
}

func getAtomFeedVersion(data io.Reader) string {
	decoder := xml_decoder.NewDecoder(data)
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		if element, ok := token.(xml.StartElement); ok {
			if element.Name.Local == "feed" {
				for _, attr := range element.Attr {
					if attr.Name.Local == "version" && attr.Value == "0.3" {
						return "0.3"
					}
				}
				return "1.0"
			}
		}
	}
	return "1.0"
}
