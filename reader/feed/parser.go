// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package feed // import "miniflux.app/reader/feed"

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"time"

	"miniflux.app/errors"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/atom"
	"miniflux.app/reader/encoding"
	"miniflux.app/reader/json"
	"miniflux.app/reader/rdf"
	"miniflux.app/reader/rss"
	"miniflux.app/timer"
)

// List of feed formats.
const (
	FormatRDF     = "rdf"
	FormatRSS     = "rss"
	FormatAtom    = "atom"
	FormatJSON    = "json"
	FormatUnknown = "unknown"
)

// DetectFeedFormat detect feed format from input data.
func DetectFeedFormat(r io.Reader) string {
	defer timer.ExecutionTime(time.Now(), "[Feed:DetectFeedFormat]")

	var buffer bytes.Buffer
	tee := io.TeeReader(r, &buffer)

	decoder := xml.NewDecoder(tee)
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

	if strings.HasPrefix(strings.TrimSpace(buffer.String()), "{") {
		return FormatJSON
	}

	return FormatUnknown
}

func parseFeed(r io.Reader) (*model.Feed, *errors.LocalizedError) {
	defer timer.ExecutionTime(time.Now(), "[Feed:ParseFeed]")

	var buffer bytes.Buffer
	size, _ := io.Copy(&buffer, r)
	if size == 0 {
		return nil, errors.NewLocalizedError("This feed is empty")
	}

	str := stripInvalidXMLCharacters(buffer.String())
	reader := strings.NewReader(str)
	format := DetectFeedFormat(reader)
	reader.Seek(0, io.SeekStart)

	switch format {
	case FormatAtom:
		return atom.Parse(reader)
	case FormatRSS:
		return rss.Parse(reader)
	case FormatJSON:
		return json.Parse(reader)
	case FormatRDF:
		return rdf.Parse(reader)
	default:
		return nil, errors.NewLocalizedError("Unsupported feed format")
	}
}

func stripInvalidXMLCharacters(input string) string {
	return strings.Map(func(r rune) rune {
		if isInCharacterRange(r) {
			return r
		}

		logger.Debug("Strip invalid XML characters: %U", r)
		return -1
	}, input)
}

// Decide whether the given rune is in the XML Character Range, per
// the Char production of http://www.xml.com/axml/testaxml.htm,
// Section 2.2 Characters.
func isInCharacterRange(r rune) (inrange bool) {
	return r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xDF77 ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}
