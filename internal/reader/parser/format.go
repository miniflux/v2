// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package parser // import "miniflux.app/v2/internal/reader/parser"

import (
	"encoding/xml"
	"io"
	"unicode"

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
	if isJSON, err := detectJSONFormat(r); err == nil && isJSON {
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

// detectJSONFormat checks if the reader contains JSON by reading until it finds
// the first non-whitespace character or reaches EOF/error.
func detectJSONFormat(r io.ReadSeeker) (bool, error) {
	const bufferSize = 32
	buffer := make([]byte, bufferSize)

	for {
		n, err := r.Read(buffer)
		if n == 0 {
			if err == io.EOF {
				return false, nil // No non-whitespace content found
			}
			return false, err
		}

		// Check each byte in the buffer
		for i := range n {
			ch := buffer[i]
			// Skip whitespace characters (space, tab, newline, carriage return, etc.)
			if unicode.IsSpace(rune(ch)) {
				continue
			}
			// First non-whitespace character determines if it's JSON
			return ch == '{', nil
		}

		// If we've read less than bufferSize, we've reached EOF
		if n < bufferSize {
			return false, nil
		}
	}
}
