// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package atom // import "miniflux.app/v2/internal/reader/atom"

import (
	"fmt"
	"io"

	"miniflux.app/v2/internal/model"
	xml_decoder "miniflux.app/v2/internal/reader/xml"
)

type atomFeed interface {
	Transform(baseURL string) *model.Feed
}

// Parse returns a normalized feed struct from a Atom feed.
func Parse(baseURL string, r io.ReadSeeker, version string) (*model.Feed, error) {
	var rawFeed atomFeed
	if version == "0.3" {
		rawFeed = new(atom03Feed)
	} else {
		rawFeed = new(atom10Feed)
	}

	if err := xml_decoder.NewXMLDecoder(r).Decode(rawFeed); err != nil {
		return nil, fmt.Errorf("atom: unable to parse feed: %w", err)
	}

	return rawFeed.Transform(baseURL), nil
}
