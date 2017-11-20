// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rss

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/miniflux/miniflux2/model"

	"golang.org/x/net/html/charset"
)

// Parse returns a normalized feed struct.
func Parse(data io.Reader) (*model.Feed, error) {
	feed := new(rssFeed)
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(feed)
	if err != nil {
		return nil, fmt.Errorf("unable to parse RSS feed: %v", err)
	}

	return feed.Transform(), nil
}
