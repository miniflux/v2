// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package feed

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/reader/atom"
	"github.com/miniflux/miniflux/reader/encoding"
	"github.com/miniflux/miniflux/reader/json"
	"github.com/miniflux/miniflux/reader/rdf"
	"github.com/miniflux/miniflux/reader/rss"
	"github.com/miniflux/miniflux/timer"
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

func parseFeed(r io.Reader) (*model.Feed, error) {
	defer timer.ExecutionTime(time.Now(), "[Feed:ParseFeed]")

	var buffer bytes.Buffer
	io.Copy(&buffer, r)

	reader := bytes.NewReader(buffer.Bytes())
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
		return nil, errors.New("Unsupported feed format")
	}
}
