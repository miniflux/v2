// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client // import "miniflux.app/http/client"

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestIsNotFound(t *testing.T) {
	scenarios := map[int]bool{
		200: false,
		404: true,
		410: true,
	}

	for input, expected := range scenarios {
		r := &Response{StatusCode: input}
		actual := r.IsNotFound()

		if actual != expected {
			t.Errorf(`Unexpected result, got %v instead of %v for status code %d`, actual, expected, input)
		}
	}
}

func TestIsNotAuthorized(t *testing.T) {
	scenarios := map[int]bool{
		200: false,
		401: true,
		403: false,
	}

	for input, expected := range scenarios {
		r := &Response{StatusCode: input}
		actual := r.IsNotAuthorized()

		if actual != expected {
			t.Errorf(`Unexpected result, got %v instead of %v for status code %d`, actual, expected, input)
		}
	}
}

func TestHasServerFailure(t *testing.T) {
	scenarios := map[int]bool{
		200: false,
		404: true,
		500: true,
	}

	for input, expected := range scenarios {
		r := &Response{StatusCode: input}
		actual := r.HasServerFailure()

		if actual != expected {
			t.Errorf(`Unexpected result, got %v instead of %v for status code %d`, actual, expected, input)
		}
	}
}

func TestIsModifiedWith304Status(t *testing.T) {
	r := &Response{StatusCode: 304}
	if r.IsModified("etag", "lastModified") {
		t.Error("The resource should not be considered modified")
	}
}

func TestIsModifiedWithIdenticalEtag(t *testing.T) {
	r := &Response{StatusCode: 200, ETag: "etag"}
	if r.IsModified("etag", "lastModified") {
		t.Error("The resource should not be considered modified")
	}
}

func TestIsModifiedWithIdenticalLastModified(t *testing.T) {
	r := &Response{StatusCode: 200, LastModified: "lastModified"}
	if r.IsModified("etag", "lastModified") {
		t.Error("The resource should not be considered modified")
	}
}

func TestIsModifiedWithDifferentHeaders(t *testing.T) {
	r := &Response{StatusCode: 200, ETag: "some etag", LastModified: "some date"}
	if !r.IsModified("etag", "lastModified") {
		t.Error("The resource should be considered modified")
	}
}

func TestToString(t *testing.T) {
	input := `test`
	r := &Response{Body: strings.NewReader(input)}

	if r.BodyAsString() != input {
		t.Error(`Unexpected output`)
	}
}

func TestEnsureUnicodeWithHTMLDocuments(t *testing.T) {
	var unicodeTestCases = []struct {
		filename, contentType string
		convertedToUnicode    bool
	}{
		{"HTTP-charset.html", "text/html; charset=iso-8859-15", true},
		{"UTF-16LE-BOM.html", "", true},
		{"UTF-16BE-BOM.html", "", true},
		{"meta-content-attribute.html", "text/html", true},
		{"meta-charset-attribute.html", "text/html", true},
		{"No-encoding-declaration.html", "text/html", true},
		{"HTTP-vs-UTF-8-BOM.html", "text/html; charset=iso-8859-15", true},
		{"HTTP-vs-meta-content.html", "text/html; charset=iso-8859-15", true},
		{"HTTP-vs-meta-charset.html", "text/html; charset=iso-8859-15", true},
		{"UTF-8-BOM-vs-meta-content.html", "text/html", true},
		{"UTF-8-BOM-vs-meta-charset.html", "text/html", true},
		{"windows_1251.html", "text/html; charset=windows-1251", true},
		{"gb2312.html", "text/html", true},
		{"urdu.xml", "text/xml; charset=utf-8", true},
		{"content-type-only-win-8859-1.xml", "application/xml; charset=ISO-8859-1", true},
		{"rdf_utf8.xml", "application/rss+xml; charset=utf-8", true},
		{"rdf_utf8.xml", "application/rss+xml; charset: utf-8", true}, // Invalid Content-Type
		{"charset-content-type-xml-iso88591.xml", "application/rss+xml; charset=ISO-8859-1", false},
		{"windows_1251.xml", "text/xml", false},
		{"smallfile.xml", "text/xml; charset=utf-8", true},
		{"single_quote_xml_encoding.xml", "text/xml; charset=utf-8", true},
	}

	for _, tc := range unicodeTestCases {
		content, err := os.ReadFile("testdata/" + tc.filename)
		if err != nil {
			t.Fatalf(`Unable to read file %q: %v`, tc.filename, err)
		}

		r := &Response{Body: bytes.NewReader(content), ContentType: tc.contentType}
		parseErr := r.EnsureUnicodeBody()
		if parseErr != nil {
			t.Fatalf(`Unicode conversion error for %q - %q: %v`, tc.filename, tc.contentType, parseErr)
		}

		isUnicode := utf8.ValidString(r.BodyAsString())
		if isUnicode != tc.convertedToUnicode {
			t.Errorf(`Unicode conversion %q - %q, got: %v, expected: %v`,
				tc.filename, tc.contentType, isUnicode, tc.convertedToUnicode)
		}
	}
}
