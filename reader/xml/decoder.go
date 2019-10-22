// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package xml // import "miniflux.app/reader/xml"

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"

	"miniflux.app/reader/encoding"
)

// NewDecoder returns a XML decoder that filters out illegal characters
func NewDecoder(data io.Reader) *xml.Decoder {
	decoder := xml.NewDecoder(data)
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		utf8Reader, err := encoding.CharsetReader(charset, input)
		if err != nil {
			return nil, err
		}
		// be more tolerant to the payload by filtering out illegal characters
		rawData, err := ioutil.ReadAll(utf8Reader)
		if err != nil {
			return nil, fmt.Errorf("Unable to read data: %q", err)
		}
		filteredBytes := bytes.Map(filterValidXMLChar, rawData)
		return bytes.NewReader(filteredBytes), nil
	}

	return decoder
}

// filterValidXMLChar is copied from encoding/xml package,
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
