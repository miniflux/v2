// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package encoding // import "miniflux.app/v2/internal/reader/encoding"

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
	"unicode/utf8"

	textencoding "golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

var htmlReaderConstructors = []struct {
	name      string
	newReader func(input []byte, contentType string) (io.Reader, error)
}{
	{
		name: "NewCharsetReader",
		newReader: func(input []byte, contentType string) (io.Reader, error) {
			return NewCharsetReader(bytes.NewReader(input), contentType)
		},
	},
	{
		name:      "NewCharsetReaderFromBytes",
		newReader: NewCharsetReaderFromBytes,
	},
}

func readFixture(tb testing.TB, name string) []byte {
	tb.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		tb.Fatalf("Unable to read fixture %q: %v", name, err)
	}

	return data
}

func encodeString(tb testing.TB, textEncoding textencoding.Encoding, text string) []byte {
	tb.Helper()

	data, err := textEncoding.NewEncoder().Bytes([]byte(text))
	if err != nil {
		tb.Fatalf("Unable to encode %q: %v", text, err)
	}

	return data
}

func readAllUTF8(t *testing.T, reader io.Reader) string {
	t.Helper()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if !utf8.Valid(data) {
		t.Fatalf("Data is not valid UTF-8: %q", data)
	}

	return string(data)
}

func assertContains(t *testing.T, data string, expectedContents []string) {
	t.Helper()

	for _, expectedContent := range expectedContents {
		if !strings.Contains(data, expectedContent) {
			t.Errorf("Data does not contain expected unicode string %q", expectedContent)
		}
	}
}

func TestCharsetReader(t *testing.T) {
	testCases := []struct {
		name             string
		label            string
		fixture          string
		expectedContents []string
	}{
		{
			name:             "UTF-8 document",
			label:            "UTF-8",
			fixture:          "utf8.xml",
			expectedContents: []string{"Café"},
		},
		{
			name:             "ISO-8859-1 document",
			label:            "ISO-8859-1",
			fixture:          "iso-8859-1.xml",
			expectedContents: []string{"Café", "Présentation"},
		},
		{
			name:             "Windows-1252 document",
			label:            "windows-1252",
			fixture:          "windows-1252.xml",
			expectedContents: []string{"Euro €"},
		},
		{
			name:             "KOI8-R document",
			label:            "KOI8-R",
			fixture:          "koi8r.xml",
			expectedContents: []string{"Пример RSS ленты", "Привет мир! Ёжик, чай, Москва, Санкт-Петербург."},
		},
		{
			name:             "unknown label with UTF-8 document",
			label:            "invalid",
			fixture:          "invalid-prolog.xml",
			expectedContents: []string{"Café"},
		},
		{
			name:             "UTF-8 document declared as ISO-8859-1",
			label:            "ISO-8859-1",
			fixture:          "utf8-incorrect-prolog.xml",
			expectedContents: []string{"Café"},
		},
		{
			name:             "UTF-8 document declared as Windows-1252",
			label:            "windows-1252",
			fixture:          "windows-1252-incorrect-prolog.xml",
			expectedContents: []string{"Euro €"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader, err := CharsetReader(testCase.label, bytes.NewReader(readFixture(t, testCase.fixture)))
			if err != nil {
				t.Fatalf("Unable to create reader: %v", err)
			}

			assertContains(t, readAllUTF8(t, reader), testCase.expectedContents)
		})
	}
}

func TestCharsetReaderRoundTrip(t *testing.T) {
	testCases := []struct {
		label        string
		textEncoding textencoding.Encoding
		text         string
	}{
		{label: "koi8-r", textEncoding: charmap.KOI8R, text: "Привет мир"},
		{label: "KOI8-R", textEncoding: charmap.KOI8R, text: "Привет мир"},
		{label: "windows-1251", textEncoding: charmap.Windows1251, text: "Съешь же ещё этих мягких французских булок"},
		{label: "iso-8859-2", textEncoding: charmap.ISO8859_2, text: "Zażółć gęślą jaźń"},
		{label: "shift_jis", textEncoding: japanese.ShiftJIS, text: "こんにちは世界"},
		{label: "euc-jp", textEncoding: japanese.EUCJP, text: "こんにちは世界"},
		{label: "gbk", textEncoding: simplifiedchinese.GBK, text: "你好世界"},
		{label: "big5", textEncoding: traditionalchinese.Big5, text: "你好世界"},
		{label: "euc-kr", textEncoding: korean.EUCKR, text: "안녕하세요 세계"},
		{label: "utf-16le", textEncoding: unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), text: "Café 😀"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			input := encodeString(t, testCase.textEncoding, testCase.text)

			reader, err := CharsetReader(testCase.label, bytes.NewReader(input))
			if err != nil {
				t.Fatalf("Unable to create reader: %v", err)
			}

			if data := readAllUTF8(t, reader); data != testCase.text {
				t.Errorf("Data does not match expected unicode string, got %q expected %q", data, testCase.text)
			}
		})
	}
}

// The XML decoder converts UTF-16 documents to UTF-8 but keeps encoding="utf-16" in the prolog.
func TestCharsetReaderWithUTF16LabelAndUTF8Input(t *testing.T) {
	input := "<rss><channel><title>Título 😀</title></channel></rss>"

	reader, err := CharsetReader("utf-16", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	if data := readAllUTF8(t, reader); data != input {
		t.Errorf("Data does not match expected unicode string, got %q expected %q", data, input)
	}
}

func TestCharsetReaderWithUnknownLabelAndNonUTF8Input(t *testing.T) {
	input := encodeString(t, charmap.Windows1252, "Euro €")

	if _, err := CharsetReader("not-a-charset", bytes.NewReader(input)); err == nil {
		t.Fatal("Expected an error for an unknown label with non UTF-8 input")
	}
}

func TestCharsetReaderWithEmptyInput(t *testing.T) {
	reader, err := CharsetReader("ISO-8859-1", bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("Unable to create reader: %v", err)
	}

	if data := readAllUTF8(t, reader); data != "" {
		t.Errorf("Expected empty data, got %q", data)
	}
}

func TestCharsetReaderWithReadError(t *testing.T) {
	expectedErr := errors.New("read failure")

	if _, err := CharsetReader("ISO-8859-1", iotest.ErrReader(expectedErr)); !errors.Is(err, expectedErr) {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestNewCharsetReader(t *testing.T) {
	windows1251Document := `<!DOCTYPE html><html><head><meta charset="windows-1251"><title>Привет мир</title></head><body><p>Съешь же ещё этих мягких французских булок</p></body></html>`
	mislabeledWindows1251Document := `<!DOCTYPE html><html><head><meta charset="iso-8859-1"><title>Привет мир</title></head><body><p>Съешь же ещё этих мягких французских булок</p></body></html>`
	shiftJISDocument := `<!DOCTYPE html><html><head><title>こんにちは世界</title></head><body><p>日本語のテキスト</p></body></html>`
	utf16Document := `<!DOCTYPE html><html><head><title>Café</title></head><body><p>Jürgen 😀</p></body></html>`

	testCases := []struct {
		name             string
		contentType      string
		input            []byte
		expectedContents []string
		expectedOutput   string
	}{
		{
			name:             "UTF-8 HTML with charset in Content-Type",
			contentType:      "text/html; charset=UTF-8",
			input:            readFixture(t, "utf8.html"),
			expectedContents: []string{"Café"},
		},
		{
			name:             "UTF-8 HTML without charset in Content-Type",
			contentType:      "text/html",
			input:            readFixture(t, "utf8.html"),
			expectedContents: []string{"Café"},
		},
		{
			name:             "UTF-8 HTML with meta tag after 1024 bytes",
			contentType:      "text/html",
			input:            readFixture(t, "utf8-meta-after-1024.html"),
			expectedContents: []string{"Café"},
		},
		{
			name:             "ISO-8859-1 HTML with charset in Content-Type",
			contentType:      "text/html; charset=ISO-8859-1",
			input:            readFixture(t, "iso-8859-1.html"),
			expectedContents: []string{"Frédéric", "Café"},
		},
		{
			name:             "ISO-8859-1 HTML with charset in meta tag",
			contentType:      "text/html",
			input:            readFixture(t, "iso-8859-1.html"),
			expectedContents: []string{"Frédéric", "Café"},
		},
		{
			// The meta tag is not detected: the Windows-1252 fallback happens to decode ISO-8859-1 text correctly.
			name:             "ISO-8859-1 HTML with meta tag after 1024 bytes",
			contentType:      "text/html",
			input:            readFixture(t, "iso-8859-1-meta-after-1024.html"),
			expectedContents: []string{"Café"},
		},
		{
			name:             "ISO-8859-1 XML without Content-Type",
			contentType:      "",
			input:            readFixture(t, "iso-8859-1.xml"),
			expectedContents: []string{"Café", "Présentation"},
		},
		{
			name:           "KOI8-R text with charset in Content-Type",
			contentType:    "text/xml; charset=koi8-r",
			input:          encodeString(t, charmap.KOI8R, "Привет мир"),
			expectedOutput: "Привет мир",
		},
		{
			name:             "KOI8-R feed with charset in Content-Type",
			contentType:      "application/rss+xml; charset=KOI8-R",
			input:            readFixture(t, "koi8r.xml"),
			expectedContents: []string{"Тестовая лента в кодировке KOI8-R", "Проверка специальных символов"},
		},
		{
			name:           "Windows-1251 HTML with charset in meta tag",
			contentType:    "text/html",
			input:          encodeString(t, charmap.Windows1251, windows1251Document),
			expectedOutput: windows1251Document,
		},
		{
			name:           "Windows-1251 HTML with charset in Content-Type and a different charset in meta tag",
			contentType:    "text/html; charset=windows-1251",
			input:          encodeString(t, charmap.Windows1251, mislabeledWindows1251Document),
			expectedOutput: mislabeledWindows1251Document,
		},
		{
			name:           "Shift_JIS HTML with charset in Content-Type",
			contentType:    "text/html; charset=shift_jis",
			input:          encodeString(t, japanese.ShiftJIS, shiftJISDocument),
			expectedOutput: shiftJISDocument,
		},
		{
			name:             "UTF-16LE HTML with byte order mark",
			contentType:      "text/html",
			input:            encodeString(t, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM), utf16Document),
			expectedContents: []string{"Café", "Jürgen 😀"},
		},
		{
			name:             "UTF-16BE HTML with byte order mark",
			contentType:      "text/html",
			input:            encodeString(t, unicode.UTF16(unicode.BigEndian, unicode.UseBOM), utf16Document),
			expectedContents: []string{"Café", "Jürgen 😀"},
		},
	}

	for _, testCase := range testCases {
		for _, constructor := range htmlReaderConstructors {
			t.Run(testCase.name+"/"+constructor.name, func(t *testing.T) {
				reader, err := constructor.newReader(testCase.input, testCase.contentType)
				if err != nil {
					t.Fatalf("Unable to create reader: %v", err)
				}

				data := readAllUTF8(t, reader)
				assertContains(t, data, testCase.expectedContents)

				if testCase.expectedOutput != "" && data != testCase.expectedOutput {
					t.Errorf("Data does not match expected unicode string, got %q expected %q", data, testCase.expectedOutput)
				}
			})
		}
	}
}

func TestNewCharsetReaderWithEmptyInput(t *testing.T) {
	for _, constructor := range htmlReaderConstructors {
		t.Run(constructor.name, func(t *testing.T) {
			reader, err := constructor.newReader(nil, "text/html")
			if err != nil {
				t.Fatalf("Unable to create reader: %v", err)
			}

			if data := readAllUTF8(t, reader); data != "" {
				t.Errorf("Expected empty data, got %q", data)
			}
		})
	}
}

func TestNewCharsetReaderWithReadError(t *testing.T) {
	expectedErr := errors.New("read failure")

	if _, err := NewCharsetReader(iotest.ErrReader(expectedErr), "text/html"); !errors.Is(err, expectedErr) {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}

func FuzzCharsetReader(f *testing.F) {
	fixtures := []struct {
		label   string
		fixture string
	}{
		{"UTF-8", "utf8.xml"},
		{"ISO-8859-1", "iso-8859-1.xml"},
		{"windows-1252", "windows-1252.xml"},
		{"KOI8-R", "koi8r.xml"},
		{"invalid", "invalid-prolog.xml"},
		{"ISO-8859-1", "utf8-incorrect-prolog.xml"},
	}
	for _, seed := range fixtures {
		f.Add(seed.label, readFixture(f, seed.fixture))
	}
	f.Add("shift_jis", encodeString(f, japanese.ShiftJIS, "こんにちは世界"))
	f.Add("utf-16le", encodeString(f, unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), "Café 😀"))
	f.Add("not-a-charset", []byte("Euro \x80"))

	f.Fuzz(func(t *testing.T, label string, input []byte) {
		reader, err := CharsetReader(label, bytes.NewReader(input))
		if err != nil {
			return
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Unable to read data: %v", err)
		}

		if !utf8.Valid(data) {
			t.Fatalf("Data is not valid UTF-8 for label %q: %q", label, data)
		}
	})
}

func FuzzNewCharsetReaderFromBytes(f *testing.F) {
	fixtures := []struct {
		contentType string
		fixture     string
	}{
		{"text/html", "utf8.html"},
		{"text/html; charset=ISO-8859-1", "iso-8859-1.html"},
		{"text/html", "iso-8859-1-meta-after-1024.html"},
		{"text/html", "utf8-meta-after-1024.html"},
		{"application/rss+xml; charset=KOI8-R", "koi8r.xml"},
		{"", "iso-8859-1.xml"},
	}
	for _, seed := range fixtures {
		f.Add(seed.contentType, readFixture(f, seed.fixture))
	}

	// Output is not checked for valid UTF-8: charset.NewReader skips decoding when the
	// first 1024 bytes look like UTF-8, even if invalid bytes come later.
	f.Fuzz(func(t *testing.T, contentType string, input []byte) {
		reader, err := NewCharsetReaderFromBytes(input, contentType)
		if err != nil {
			t.Fatalf("Unable to create reader: %v", err)
		}

		if _, err := io.ReadAll(reader); err != nil {
			t.Fatalf("Unable to read data: %v", err)
		}
	})
}
