// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package xml // import "miniflux.app/v2/internal/reader/xml"

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"miniflux.app/v2/internal/reader/encoding"
)

// NewXMLDecoder returns a XML decoder that filters illegal characters.
func NewXMLDecoder(data io.ReadSeeker) *xml.Decoder {
	var decoder *xml.Decoder

	// This is way fasted than io.ReadAll(data) as the buffer can be allocated in one go instead of dynamically grown.
	buffer := &bytes.Buffer{}
	io.Copy(buffer, data)

	enc := getEncoding(buffer.Bytes())
	if enc == "" || strings.EqualFold(enc, "utf-8") {
		// filter invalid chars now, since decoder.CharsetReader not called for utf-8 content
		filteredBytes := bytes.Map(filterValidXMLChar, buffer.Bytes())
		decoder = xml.NewDecoder(bytes.NewReader(filteredBytes))
	} else {
		// filter invalid chars later within decoder.CharsetReader
		data.Seek(0, io.SeekStart)
		decoder = xml.NewDecoder(data)
	}

	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		utf8Reader, err := encoding.CharsetReader(charset, input)
		if err != nil {
			return nil, err
		}
		rawData, err := io.ReadAll(utf8Reader)
		if err != nil {
			return nil, fmt.Errorf("encoding: unable to read data: %w", err)
		}
		filteredBytes := bytes.Map(filterValidXMLChar, rawData)
		return bytes.NewReader(filteredBytes), nil
	}

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

// This function is copied from encoding/xml's procInst and adapted for []bytes instead of string
func getEncoding(b []byte) string {
	// TODO: this parsing is somewhat lame and not exact.
	// It works for all actual cases, though.
	idx := bytes.Index(b, []byte("encoding="))
	if idx == -1 {
		return ""
	}
	v := b[idx+len("encoding="):]
	if len(v) == 0 {
		return ""
	}
	if v[0] != '\'' && v[0] != '"' {
		return ""
	}
	idx = bytes.IndexRune(v[1:], rune(v[0]))
	if idx == -1 {
		return ""
	}
	return string(v[1 : idx+1])
}
