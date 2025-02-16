// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package encoding // import "miniflux.app/v2/internal/reader/encoding"

import (
	"bytes"
	"io"
	"os"
	"testing"
	"unicode/utf8"
)

func TestCharsetReaderWithUTF8(t *testing.T) {
	file := "testdata/utf8.xml"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := CharsetReader("UTF-8", f)
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestCharsetReaderWithISO88591(t *testing.T) {
	file := "testdata/iso-8859-1.xml"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := CharsetReader("ISO-8859-1", f)
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestCharsetReaderWithWindows1252(t *testing.T) {
	file := "testdata/windows-1252.xml"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := CharsetReader("windows-1252", f)
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Euro €"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestCharsetReaderWithInvalidProlog(t *testing.T) {
	file := "testdata/invalid-prolog.xml"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := CharsetReader("invalid", f)
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestCharsetReaderWithUTF8DocumentWithIncorrectProlog(t *testing.T) {
	file := "testdata/utf8-incorrect-prolog.xml"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := CharsetReader("ISO-8859-1", f)
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestCharsetReaderWithWindows1252DocumentWithIncorrectProlog(t *testing.T) {
	file := "testdata/windows-1252-incorrect-prolog.xml"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := CharsetReader("windows-1252", f)
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Euro €"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestNewReaderWithUTF8Document(t *testing.T) {
	file := "testdata/utf8.html"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := NewCharsetReader(f, "text/html; charset=UTF-8")
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestNewReaderWithUTF8DocumentAndNoContentEncoding(t *testing.T) {
	file := "testdata/utf8.html"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := NewCharsetReader(f, "text/html")
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestNewReaderWithISO88591Document(t *testing.T) {
	file := "testdata/iso-8859-1.xml"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := NewCharsetReader(f, "text/html; charset=ISO-8859-1")
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestNewReaderWithISO88591DocumentAndNoContentType(t *testing.T) {
	file := "testdata/iso-8859-1.xml"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := NewCharsetReader(f, "")
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestNewReaderWithISO88591DocumentWithMetaAfter1024Bytes(t *testing.T) {
	file := "testdata/iso-8859-1-meta-after-1024.html"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := NewCharsetReader(f, "text/html")
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}

func TestNewReaderWithUTF8DocumentWithMetaAfter1024Bytes(t *testing.T) {
	file := "testdata/utf8-meta-after-1024.html"

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}

	reader, err := NewCharsetReader(f, "text/html")
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8")
	}

	expectedUnicodeString := "Café"
	if !bytes.Contains(data, []byte(expectedUnicodeString)) {
		t.Fatalf("Data does not contain expected unicode string: %s", expectedUnicodeString)
	}
}
