// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package xml // import "miniflux.app/v2/internal/reader/xml"

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestXMLDocumentWithISO88591Encoding(t *testing.T) {
	fp, err := os.Open("testdata/iso88591.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()

	type myXMLDocument struct {
		XMLName xml.Name `xml:"note"`
		To      string   `xml:"to"`
		From    string   `xml:"from"`
	}

	var doc myXMLDocument

	decoder := NewXMLDecoder(fp)
	err = decoder.Decode(&doc)
	if err != nil {
		t.Fatal(err)
	}

	expectedTo := "Anaïs"
	expectedFrom := "Jürgen"

	if doc.To != expectedTo {
		t.Errorf(`Incorrect "to" field, expected: %q, got: %q`, expectedTo, doc.To)
	}
	if doc.From != expectedFrom {
		t.Errorf(`Incorrect "from" field, expected: %q, got: %q`, expectedFrom, doc.From)
	}
}

func TestXMLDocumentWithISO88591FileEncodingButUTF8Prolog(t *testing.T) {
	fp, err := os.Open("testdata/iso88591_utf8_mismatch.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()

	type myXMLDocument struct {
		XMLName xml.Name `xml:"note"`
		To      string   `xml:"to"`
		From    string   `xml:"from"`
	}

	var doc myXMLDocument

	decoder := NewXMLDecoder(fp)
	err = decoder.Decode(&doc)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: detect actual encoding from bytes if not UTF-8 and convert to UTF-8 if needed.
	// For now we just expect the invalid characters to be stripped out.
	expectedTo := "Anas"
	expectedFrom := "Jrgen"

	if doc.To != expectedTo {
		t.Errorf(`Incorrect "to" field, expected: %q, got: %q`, expectedTo, doc.To)
	}
	if doc.From != expectedFrom {
		t.Errorf(`Incorrect "from" field, expected: %q, got: %q`, expectedFrom, doc.From)
	}
}

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
