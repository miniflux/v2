// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package encoding // import "miniflux.app/v2/internal/reader/encoding"

import (
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
}
