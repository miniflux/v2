// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package parser // import "miniflux.app/v2/internal/reader/parser"

import (
	"bytes"
	"encoding/xml"
	"io"

	rxml "miniflux.app/v2/internal/reader/xml"
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
func DetectFeedFormat(r io.ReadSeeker) (string, string) {
	data := make([]byte, 512)
	r.Read(data)

	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("{")) {
		return FormatJSON, ""
	}

	r.Seek(0, io.SeekStart)
	decoder := rxml.NewXMLDecoder(r)

	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		if element, ok := token.(xml.StartElement); ok {
			switch element.Name.Local {
			case "rss":
				return FormatRSS, ""
			case "feed":
				for _, attr := range element.Attr {
					if attr.Name.Local == "version" && attr.Value == "0.3" {
						return FormatAtom, "0.3"
					}
				}
				return FormatAtom, "1.0"
			case "RDF":
				return FormatRDF, ""
			}
		}
	}

	return FormatUnknown, ""
}
