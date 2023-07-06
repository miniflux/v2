// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package icon // import "miniflux.app/reader/icon"

import (
	"strings"
	"testing"
)

func TestParseImageDataURL(t *testing.T) {
	iconURL := "data:image/webp;base64,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"
	icon, err := parseImageDataURL(iconURL)
	if err != nil {
		t.Fatalf(`We should be able to parse valid data URL: %v`, err)
	}

	if icon.MimeType != "image/webp" {
		t.Fatal(`Invalid mime type parsed`)
	}

	if icon.Hash == "" {
		t.Fatal(`Image hash should be computed`)
	}
}

func TestParseImageDataURLWithNoEncoding(t *testing.T) {
	iconURL := `data:image/webp,%3Ch1%3EHello%2C%20World%21%3C%2Fh1%3E`
	icon, err := parseImageDataURL(iconURL)
	if err != nil {
		t.Fatalf(`We should be able to parse valid data URL: %v`, err)
	}

	if icon.MimeType != "image/webp" {
		t.Fatal(`Invalid mime type parsed`)
	}

	if string(icon.Content) == "Hello, World!" {
		t.Fatal(`Value should be URL-decoded`)
	}

	if icon.Hash == "" {
		t.Fatal(`Image hash should be computed`)
	}
}

func TestParseImageDataURLWithNoMediaTypeAndNoEncoding(t *testing.T) {
	iconURL := `data:,Hello%2C%20World%21`
	_, err := parseImageDataURL(iconURL)
	if err == nil {
		t.Fatal(`We should detect invalid mime type`)
	}
}

func TestParseInvalidImageDataURLWithBadMimeType(t *testing.T) {
	_, err := parseImageDataURL("data:text/plain;base64,blob")
	if err == nil {
		t.Fatal(`We should detect invalid mime type`)
	}
}

func TestParseInvalidImageDataURLWithUnsupportedEncoding(t *testing.T) {
	_, err := parseImageDataURL("data:image/png;base32,blob")
	if err == nil {
		t.Fatal(`We should detect unsupported encoding`)
	}
}

func TestParseInvalidImageDataURLWithNoData(t *testing.T) {
	_, err := parseImageDataURL("data:image/png;base64,")
	if err == nil {
		t.Fatal(`We should detect invalid encoded data`)
	}
}

func TestParseInvalidImageDataURL(t *testing.T) {
	_, err := parseImageDataURL("data:image/jpeg")
	if err == nil {
		t.Fatal(`We should detect malformed image data URL`)
	}
}

func TestParseInvalidImageDataURLWithWrongPrefix(t *testing.T) {
	_, err := parseImageDataURL("data,test")
	if err == nil {
		t.Fatal(`We should detect malformed image data URL`)
	}
}

func TestParseDocumentWithWhitespaceIconURL(t *testing.T) {
	html := `<link rel="shortcut icon" href="
		/static/img/favicon.ico
	">`

	iconURL, err := parseDocument("http://www.example.org/", strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if iconURL != "http://www.example.org/static/img/favicon.ico" {
		t.Errorf(`Invalid icon URL, got %q`, iconURL)
	}
}
