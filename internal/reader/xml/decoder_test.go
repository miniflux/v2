// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package xml // import "miniflux.app/v2/internal/reader/xml"

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestXMLDocumentWithIllegalUnicodeCharacters(t *testing.T) {
	type myxml struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Title   string   `xml:"title"`
	}

	expected := "Title & 中文标题"
	data := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><title>Title & 中文%s标题</title></rss>`, "\x10")
	reader := strings.NewReader(data)

	var x myxml

	decoder := NewXMLDecoder(reader)
	err := decoder.Decode(&x)
	if err != nil {
		t.Error(err)
		return
	}
	if x.Title != expected {
		t.Errorf("Incorrect entry title, expected: %s, got: %s", expected, x.Title)
	}
}

func TestXMLDocumentWindows251EncodedWithIllegalCharacters(t *testing.T) {
	type myxml struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Title   string   `xml:"title"`
	}

	expected := "Title & 中文标题"
	data := fmt.Sprintf(`<?xml version="1.0" encoding="windows-1251"?><rss version="2.0"><title>Title & 中文%s标题</title></rss>`, "\x10")
	reader := strings.NewReader(data)

	var x myxml

	decoder := NewXMLDecoder(reader)
	err := decoder.Decode(&x)
	if err != nil {
		t.Error(err)
		return
	}
	if x.Title != expected {
		t.Errorf("Incorrect entry title, expected: %s, got: %s", expected, x.Title)
	}
}

func TestXMLDocumentWithIncorrectEncodingField(t *testing.T) {
	type myxml struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Title   string   `xml:"title"`
	}

	expected := "Title & 中文标题"
	data := fmt.Sprintf(`<?xml version="1.0" encoding="invalid"?><rss version="2.0"><title>Title & 中文%s标题</title></rss>`, "\x10")
	reader := strings.NewReader(data)

	var x myxml

	decoder := NewXMLDecoder(reader)
	err := decoder.Decode(&x)
	if err != nil {
		t.Error(err)
		return
	}
	if x.Title != expected {
		t.Errorf("Incorrect entry title, expected: %s, got: %s", expected, x.Title)
	}
}

func TestFilterValidXMLCharsWithInvalidUTF8Sequence(t *testing.T) {
	// Create input with invalid UTF-8 sequence
	input := []byte{0x41, 0xC0, 0xAF, 0x42} // 'A', invalid UTF-8, 'B'

	filtered := filterValidXMLChars(input)

	// The function would replace invalid UTF-8 with replacement char
	// rather than properly filtering
	if utf8.Valid(filtered) {
		r, _ := utf8.DecodeRune(filtered[1:])
		if r == utf8.RuneError {
			t.Error("Invalid UTF-8 was not properly filtered")
		}
	}
}

func FuzzFilterValidXMLChars(f *testing.F) {
	f.Fuzz(func(t *testing.T, s []byte) {
		filterValidXMLChars(s)
	})
}
