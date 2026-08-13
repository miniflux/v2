// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package xml // import "miniflux.app/v2/internal/reader/xml"

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"unicode/utf8"

	"miniflux.app/v2/internal/reader/encoding"
)

// NewXMLDecoder returns a XML decoder that filters illegal characters.
func NewXMLDecoder(data io.ReadSeeker) *xml.Decoder {
	var decoder *xml.Decoder

	buffer := &bytes.Buffer{}
	if hasUTF8XMLDeclaration(data, buffer) {
		// TODO: detect actual encoding from bytes if not UTF-8 and convert to UTF-8 if needed.
		// For now we just expect the invalid characters to be stripped out.

		// The whole document is needed to filter invalid chars now, since
		// decoder.CharsetReader isn't called for utf-8 content.
		io.Copy(buffer, data)

		filteredBytes := filterValidXMLChars(buffer.Bytes())

		decoder = xml.NewDecoder(bytes.NewReader(filteredBytes))
	} else {
		data.Seek(0, io.SeekStart)
		decoder = xml.NewDecoder(data)

		// The XML document will be converted to UTF-8 by encoding.CharsetReader
		// Invalid characters will be filtered later via decoder.CharsetReader
		decoder.CharsetReader = charsetReaderFilterInvalidUtf8
	}

	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false

	return decoder
}

func charsetReaderFilterInvalidUtf8(charset string, input io.Reader) (io.Reader, error) {
	utf8Reader, err := encoding.CharsetReader(charset, input)
	if err != nil {
		return nil, err
	}
	rawData, err := io.ReadAll(utf8Reader)
	if err != nil {
		return nil, fmt.Errorf("xml: unable to read data: %w", err)
	}
	filteredBytes := filterValidXMLChars(rawData)
	return bytes.NewReader(filteredBytes), nil
}

// filterValidXMLChars filters inplace invalid XML characters.
// This function is inspired from bytes.Map
func filterValidXMLChars(s []byte) []byte {
	var i uint // declaring it as an uint removes a bound check in the loop.
	var j int

	for i = 0; i < uint(len(s)); {
		wid := 1
		r := rune(s[i])
		if r >= utf8.RuneSelf {
			r, wid = utf8.DecodeRune(s[i:])
		}
		if r != utf8.RuneError {
			if r = filterValidXMLChar(r); r >= 0 {
				utf8.EncodeRune(s[j:], r)
				j += wid
			}
		}
		i += uint(wid)
	}
	return s[:j]
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
func getEncoding(b []byte) []byte {
	// This parsing is somewhat lame and not exact.
	// It works for all actual cases, though.
	idx := bytes.Index(b, []byte("encoding="))
	if idx == -1 {
		return nil
	}
	v := b[idx+len("encoding="):]
	if len(v) == 0 {
		return nil
	}
	if v[0] != '\'' && v[0] != '"' {
		return nil
	}
	idx = bytes.IndexRune(v[1:], rune(v[0]))
	if idx == -1 {
		return nil
	}
	return v[1 : idx+1]
}

// hasUTF8XMLDeclaration reads data chunk by chunk into buffer until the XML
// declaration's encoding can be resolved, then reports whether it is UTF-8.
// The encoding sits in the XML declaration at the start of the document, but
// pathological documents may pad it with a lot of whitespace, so buffering the
// whole document up-front would be wasted work in the non-UTF-8 case.
func hasUTF8XMLDeclaration(data io.Reader, buffer *bytes.Buffer) bool {
	chunk := make([]byte, 1024)
	var enc []byte
	for {
		n, err := data.Read(chunk)
		if n > 0 {
			buffer.Write(chunk[:n])
		}
		enc = getEncoding(buffer.Bytes())
		// Stop once a complete encoding value or the end of the XML declaration
		// ("?>") is seen, after which no encoding declaration can appear.
		if err != nil || enc != nil || bytes.Contains(buffer.Bytes(), []byte("?>")) {
			break
		}
	}

	return enc == nil || bytes.EqualFold(enc, []byte("utf-8"))
}
