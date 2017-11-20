// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package feed

import (
	"bytes"
	"encoding/xml"
	"errors"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/feed/atom"
	"github.com/miniflux/miniflux2/reader/feed/json"
	"github.com/miniflux/miniflux2/reader/feed/rss"
	"io"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

const (
	FormatRss     = "rss"
	FormatAtom    = "atom"
	FormatJson    = "json"
	FormatUnknown = "unknown"
)

func DetectFeedFormat(data io.Reader) string {
	defer helper.ExecutionTime(time.Now(), "[Feed:DetectFeedFormat]")

	var buffer bytes.Buffer
	tee := io.TeeReader(data, &buffer)

	decoder := xml.NewDecoder(tee)
	decoder.CharsetReader = charset.NewReaderLabel

	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		if element, ok := token.(xml.StartElement); ok {
			switch element.Name.Local {
			case "rss":
				return FormatRss
			case "feed":
				return FormatAtom
			}
		}
	}

	if strings.HasPrefix(strings.TrimSpace(buffer.String()), "{") {
		return FormatJson
	}

	return FormatUnknown
}

func parseFeed(data io.Reader) (*model.Feed, error) {
	defer helper.ExecutionTime(time.Now(), "[Feed:ParseFeed]")

	var buffer bytes.Buffer
	io.Copy(&buffer, data)

	reader := bytes.NewReader(buffer.Bytes())
	format := DetectFeedFormat(reader)
	reader.Seek(0, io.SeekStart)

	switch format {
	case FormatAtom:
		return atom.Parse(reader)
	case FormatRss:
		return rss.Parse(reader)
	case FormatJson:
		return json.Parse(reader)
	default:
		return nil, errors.New("Unsupported feed format")
	}
}
