// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package parser // import "miniflux.app/reader/parser"

import (
	"strings"

	"miniflux.app/errors"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/atom"
	"miniflux.app/reader/json"
	"miniflux.app/reader/rdf"
	"miniflux.app/reader/rss"
)

// ParseFeed analyzes the input data and returns a normalized feed object.
func ParseFeed(data string) (*model.Feed, *errors.LocalizedError) {
	data = stripInvalidXMLCharacters(data)

	switch DetectFeedFormat(data) {
	case FormatAtom:
		return atom.Parse(strings.NewReader(data))
	case FormatRSS:
		return rss.Parse(strings.NewReader(data))
	case FormatJSON:
		return json.Parse(strings.NewReader(data))
	case FormatRDF:
		return rdf.Parse(strings.NewReader(data))
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
