// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package icon // import "miniflux.app/v2/internal/reader/icon"

import (
	"bytes"
	"encoding/base64"
	"image"
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
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

func TestParseImageWithRawSVGEncodedInUTF8(t *testing.T) {
	iconURL := `data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 456 456'><circle></circle></svg>`
	icon, err := parseImageDataURL(iconURL)
	if err != nil {
		t.Fatalf(`We should be able to parse valid data URL: %v`, err)
	}

	if icon.MimeType != "image/svg+xml" {
		t.Fatal(`Invalid mime type parsed`)
	}

	if icon.Hash == "" {
		t.Fatal(`Image hash should be computed`)
	}

	if string(icon.Content) != `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 456 456'><circle></circle></svg>` {
		t.Fatal(`Invalid SVG content`)
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

func TestFindIconURLsFromHTMLDocument_MultipleIcons(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
	<link rel="icon" href="/favicon.ico">
	<link rel="shortcut icon" href="/shortcut-favicon.ico">
	<link rel="icon shortcut" href="/icon-shortcut.ico">
	<link rel="apple-touch-icon" href="/apple-touch-icon.png">
</head>
</html>`

	iconURLs, err := findIconURLsFromHTMLDocument("https://example.org", strings.NewReader(html), "text/html")
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"https://example.org/favicon.ico",
		"https://example.org/shortcut-favicon.ico",
		"https://example.org/icon-shortcut.ico",
		"https://example.org/apple-touch-icon.png",
	}

	if len(iconURLs) != len(expected) {
		t.Fatalf("Expected %d icon URLs, got %d", len(expected), len(iconURLs))
	}

	for i, expectedURL := range expected {
		if iconURLs[i] != expectedURL {
			t.Errorf("Expected icon URL %d to be %q, got %q", i, expectedURL, iconURLs[i])
		}
	}
}

func TestFindIconURLsFromHTMLDocument_CaseInsensitiveRel(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
	<link rel="ICON" href="/favicon1.ico">
	<link rel="Icon" href="/favicon2.ico">
	<link rel="SHORTCUT ICON" href="/favicon3.ico">
	<link rel="Shortcut Icon" href="/favicon4.ico">
	<link rel="ICON SHORTCUT" href="/favicon5.ico">
	<link rel="Icon Shortcut" href="favicon6.ico">
</head>
</html>`

	iconURLs, err := findIconURLsFromHTMLDocument("https://example.org/folder/", strings.NewReader(html), "text/html")
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"https://example.org/favicon1.ico",
		"https://example.org/favicon2.ico",
		"https://example.org/favicon3.ico",
		"https://example.org/favicon4.ico",
		"https://example.org/favicon5.ico",
		"https://example.org/folder/favicon6.ico",
	}

	if len(iconURLs) != len(expected) {
		t.Fatalf("Expected %d icon URLs, got %d", len(expected), len(iconURLs))
	}

	for i, expectedURL := range expected {
		if iconURLs[i] != expectedURL {
			t.Errorf("Expected icon URL %d to be %q, got %q", i, expectedURL, iconURLs[i])
		}
	}
}

func TestFindIconURLsFromHTMLDocument_NoIcons(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>No Icons Here</title>
	<link rel="stylesheet" href="/style.css">
	<link rel="canonical" href="https://example.com">
</head>
</html>`

	iconURLs, err := findIconURLsFromHTMLDocument("https://example.org", strings.NewReader(html), "text/html")
	if err != nil {
		t.Fatal(err)
	}

	if len(iconURLs) != 0 {
		t.Fatalf("Expected 0 icon URLs, got %d: %v", len(iconURLs), iconURLs)
	}
}

func TestFindIconURLsFromHTMLDocument_EmptyHref(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
	<link rel="icon" href="">
	<link rel="icon" href="   ">
	<link rel="icon">
	<link rel="shortcut icon" href="/valid-icon.ico">
</head>
</html>`

	iconURLs, err := findIconURLsFromHTMLDocument("https://example.org", strings.NewReader(html), "text/html")
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"https://example.org/valid-icon.ico"}

	if len(iconURLs) != len(expected) {
		t.Fatalf("Expected %d icon URLs, got %d", len(expected), len(iconURLs))
	}

	if iconURLs[0] != expected[0] {
		t.Errorf("Expected icon URL to be %q, got %q", expected[0], iconURLs[0])
	}
}

func TestFindIconURLsFromHTMLDocument_DataURLs(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
	<link rel="icon" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAGAhGAQ+QAAAABJRU5ErkJggg==">
	<link rel="shortcut icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg'></svg>">
	<link rel="icon" href="/regular-icon.ico">
</head>
</html>`

	iconURLs, err := findIconURLsFromHTMLDocument("https://example.org/folder", strings.NewReader(html), "text/html")
	if err != nil {
		t.Fatal(err)
	}

	// The function processes queries in order: rel="icon", then rel="shortcut icon", etc.
	// So both rel="icon" links are found first, then the rel="shortcut icon" link
	expected := []string{
		"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAGAhGAQ+QAAAABJRU5ErkJggg==",
		"https://example.org/regular-icon.ico",
		"data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg'></svg>",
	}

	if len(iconURLs) != len(expected) {
		t.Fatalf("Expected %d icon URLs, got %d", len(expected), len(iconURLs))
	}

	for i, expectedURL := range expected {
		if iconURLs[i] != expectedURL {
			t.Errorf("Expected icon URL %d to be %q, got %q", i, expectedURL, iconURLs[i])
		}
	}
}

func TestFindIconURLsFromHTMLDocument_RelativeAndAbsoluteURLs(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
	<link rel="icon" href="/absolute-path.ico">
	<link rel="icon" href="relative-path.ico">
	<link rel="icon" href="../parent-dir.ico">
	<link rel="icon" href="https://example.com/external.ico">
	<link rel="icon" href="//cdn.example.com/protocol-relative.ico">
</head>
</html>`

	iconURLs, err := findIconURLsFromHTMLDocument("https://example.org/folder/", strings.NewReader(html), "text/html")
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"https://example.org/absolute-path.ico",
		"https://example.org/folder/relative-path.ico",
		"https://example.org/parent-dir.ico",
		"https://example.com/external.ico",
		"https://cdn.example.com/protocol-relative.ico",
	}

	if len(iconURLs) != len(expected) {
		t.Fatalf("Expected %d icon URLs, got %d", len(expected), len(iconURLs))
	}

	for i, expectedURL := range expected {
		if iconURLs[i] != expectedURL {
			t.Errorf("Expected icon URL %d to be %q, got %q", i, expectedURL, iconURLs[i])
		}
	}
}

func TestFindIconURLsFromHTMLDocument_InvalidHTML(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
	<link rel="icon" href="/valid-before-error.ico">
	<link rel="icon" href="/unclosed-tag.ico"
	<link rel="shortcut icon" href="/valid-after-error.ico">
</head>
</html>`

	iconURLs, err := findIconURLsFromHTMLDocument("https://example.org", strings.NewReader(html), "text/html")
	if err != nil {
		t.Fatal(err)
	}

	// goquery should handle malformed HTML gracefully
	if len(iconURLs) == 0 {
		t.Fatal("Expected to find some icon URLs even with malformed HTML")
	}

	// Should at least find the valid ones
	foundValidIcon := false
	for _, url := range iconURLs {
		if url == "https://example.org/valid-before-error.ico" || url == "https://example.org/valid-after-error.ico" {
			foundValidIcon = true
			break
		}
	}

	if !foundValidIcon {
		t.Errorf("Expected to find at least one valid icon URL, got: %v", iconURLs)
	}
}

func TestFindIconURLsFromHTMLDocument_EmptyDocument(t *testing.T) {
	iconURLs, err := findIconURLsFromHTMLDocument("https://example.org", strings.NewReader(""), "text/html")
	if err != nil {
		t.Fatal(err)
	}

	if len(iconURLs) != 0 {
		t.Fatalf("Expected 0 icon URLs from empty document, got %d", len(iconURLs))
	}
}

func TestResizeIconSmallGif(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString("R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==")
	if err != nil {
		t.Fatal(err)
	}
	icon := model.Icon{
		Content:  data,
		MimeType: "image/gif",
	}
	if !bytes.Equal(icon.Content, resizeIcon(&icon).Content) {
		t.Fatalf("Converted gif smaller than 16x16")
	}
}

func TestResizeIconPng(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAACEAAAAhCAYAAABX5MJvAAAALUlEQVR42u3OMQEAAAgDoJnc6BpjDyRgcrcpGwkJCQkJCQkJCQkJCQkJCYmyB7NfUj/Kk4FkAAAAAElFTkSuQmCC")
	if err != nil {
		t.Fatal(err)
	}
	icon := model.Icon{
		Content:  data,
		MimeType: "image/png",
	}
	resizedIcon := resizeIcon(&icon)

	if bytes.Equal(data, resizedIcon.Content) {
		t.Fatalf("Didn't convert png of 33x33")
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(resizedIcon.Content))
	if err != nil {
		t.Fatalf("Couln't decode resulting png: %v", err)
	}

	if config.Height != 32 || config.Width != 32 {
		t.Fatalf("Was expecting an image of 16x16, got %dx%d", config.Width, config.Height)
	}
}

func TestResizeIconWebp(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString("UklGRkAAAABXRUJQVlA4IDQAAADwAQCdASoBAAEAAQAcJaACdLoB+AAETAAA/vW4f/6aR40jxpHxcP/ugT90CfugT/3NoAAA")
	if err != nil {
		t.Fatal(err)
	}
	icon := model.Icon{
		Content:  data,
		MimeType: "image/webp",
	}

	if !bytes.Equal(icon.Content, resizeIcon(&icon).Content) {
		t.Fatalf("Converted webp smaller than 16x16")
	}
}

func TestEnsureRemoteIconURLAllowedRejectsPrivateNetworks(t *testing.T) {
	if err := ensureRemoteIconURLAllowed("http://192.168.0.1/favicon.ico", false); err == nil {
		t.Fatal("Expected private network hosts to be rejected")
	}
}

func TestEnsureRemoteIconURLAllowedAllowsPublicNetworks(t *testing.T) {
	if err := ensureRemoteIconURLAllowed("https://1.1.1.1/favicon.ico", false); err != nil {
		t.Fatalf("Expected public network hosts to be allowed: %v", err)
	}
}

func TestEnsureRemoteIconURLAllowedAllowsPrivateWhenEnabled(t *testing.T) {
	if err := ensureRemoteIconURLAllowed("http://10.0.0.5/icon.png", true); err != nil {
		t.Fatalf("Expected private network hosts to be allowed when explicitly enabled: %v", err)
	}
}

func TestResizeInvalidImage(t *testing.T) {
	icon := model.Icon{
		Content:  []byte("invalid data"),
		MimeType: "image/gif",
	}
	if !bytes.Equal(icon.Content, resizeIcon(&icon).Content) {
		t.Fatalf("Tried to convert an invalid image")
	}
}

func TestMinifySvg(t *testing.T) {
	data := []byte(`<svg path d=" M1 4h-.001 V1h2v.001 M1 2.6 h1v.001"/></svg>`)
	want := []byte(`<svg path="" d="M1 4H.999V1h2v.001M1 2.6h1v.001"/></svg>`)
	icon := model.Icon{Content: data, MimeType: "image/svg+xml"}
	got := resizeIcon(&icon).Content
	if !bytes.Equal(want, got) {
		t.Fatalf("Didn't correctly minify the svg: got %s instead of %s", got, want)
	}
}

func TestMinifySvgWithError(t *testing.T) {
	// Invalid SVG with malformed XML that should cause minification to fail
	data := []byte(`<svg><><invalid-tag<>unclosed`)
	original := make([]byte, len(data))
	copy(original, data)

	icon := model.Icon{
		Content:  data,
		MimeType: "image/svg+xml",
	}

	result := resizeIcon(&icon)

	// When minification fails, the original content should be preserved
	if !bytes.Equal(original, result.Content) {
		t.Fatalf("Expected original content to be preserved on minification error, got %s instead of %s", result.Content, original)
	}

	// MimeType should remain unchanged
	if result.MimeType != "image/svg+xml" {
		t.Fatalf("Expected MimeType to remain image/svg+xml, got %s", result.MimeType)
	}
}
