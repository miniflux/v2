// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rss

import (
	"encoding/xml"
	"fmt"
	"github.com/miniflux/miniflux2/model"
	"io"

	"golang.org/x/net/html/charset"
)

// Parse returns a normalized feed struct.
func Parse(data io.Reader) (*model.Feed, error) {
	rssFeed := new(RssFeed)
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(rssFeed)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse RSS feed: %v", err)
	}

	return rssFeed.Transform(), nil
}
