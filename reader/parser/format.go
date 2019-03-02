// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package parser // import "miniflux.app/reader/parser"

import (
	"encoding/xml"
	"strings"

	"miniflux.app/reader/encoding"
)

// List of feed formats.
const (
	FormatRDF     = "rdf"
	FormatRSS     = "rss"
	FormatAtom    = "atom"
	FormatJSON    = "json"
	FormatUnknown = "unknown"
)

// DetectFeedFormat tries to guess the feed format from input data.
func DetectFeedFormat(data string) string {
	if strings.HasPrefix(strings.TrimSpace(data), "{") {
		return FormatJSON
	}

	decoder := xml.NewDecoder(strings.NewReader(data))
	decoder.Entity = xml.HTMLEntity
	decoder.CharsetReader = encoding.CharsetReader

	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		if element, ok := token.(xml.StartElement); ok {
			switch element.Name.Local {
			case "rss":
				return FormatRSS
			case "feed":
				return FormatAtom
			case "RDF":
				return FormatRDF
			}
		}
	}

	return FormatUnknown
}
