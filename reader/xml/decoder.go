// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package xml // import "miniflux.app/reader/xml"

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"strings"
)

// NewDecoder returns a XML decoder that filters illegal characters.
func NewDecoder(data io.Reader) *xml.Decoder {
	var newr io.Reader
	rawData, err := ioutil.ReadAll(data)
	if err == nil {
		xmlStr := strings.Map(filterValidXMLChar, string(rawData))
		newr = strings.NewReader(xmlStr)
	} else {
		newr = data
	}

	decoder := xml.NewDecoder(newr)
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false

	return decoder
}

// This function is copied from encoding/xml package,
// and is used to check if all the characters are legal.
func filterValidXMLChar(r rune) rune {
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
