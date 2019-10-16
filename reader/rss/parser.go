// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rss // import "miniflux.app/reader/rss"

import (
	"bytes"
	"encoding/xml"
	"io"
	"io/ioutil"

	"miniflux.app/errors"
	"miniflux.app/model"
	"miniflux.app/reader/encoding"
)

// Parse returns a normalized feed struct from a RSS feed.
func Parse(data io.Reader) (*model.Feed, *errors.LocalizedError) {
	feed := new(rssFeed)
	decoder := xml.NewDecoder(data)
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		utf8Reader, err := encoding.CharsetReader(charset, input)
		if err != nil {
			return nil, err
		}
		// be more tolerant to the payload by filtering illegal characters
		rawData, err := ioutil.ReadAll(utf8Reader)
		if err != nil {
			return nil, errors.NewLocalizedError("Unable to read data: %q", err)
		}
		filteredBytes := bytes.Map(isInCharacterRange, rawData)
		return bytes.NewReader(filteredBytes), nil
	}

	err := decoder.Decode(feed)
	if err != nil {
		return nil, errors.NewLocalizedError("Unable to parse RSS feed: %q", err)
	}

	return feed.Transform(), nil
}

// isInCharacterRange is copied from encoding/xml package,
// and is used to check if all the characters are legal.
func isInCharacterRange(r rune) rune {
	if r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xD7FF ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF {
		return r
	}
	return -1
}
