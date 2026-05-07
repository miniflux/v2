// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package scraper // import "miniflux.app/v2/internal/reader/scraper"

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractHeadMetadataOpenGraph(t *testing.T) {
	htmlInput := `<!doctype html>
<html><head>
<meta property="og:title" content="A picture of a cat">
<meta property="og:description" content="I took this picture of a cat today.">
<meta property="og:image" content="https://cdn.example.org/cat.jpg">
<meta property="og:url" content="https://example.org/posts/1">
</head><body>ignored</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("unable to parse fixture: %v", err)
	}

	metadata := ExtractHeadMetadata(doc)

	expected := map[string]string{
		"og:title":       "A picture of a cat",
		"og:description": "I took this picture of a cat today.",
		"og:image":       "https://cdn.example.org/cat.jpg",
		"og:url":         "https://example.org/posts/1",
	}

	if len(metadata) != len(expected) {
		t.Fatalf("expected %d entries, got %d (%v)", len(expected), len(metadata), metadata)
	}

	for k, v := range expected {
		if metadata[k] != v {
			t.Errorf("metadata[%q] = %q, want %q", k, metadata[k], v)
		}
	}
}

func TestExtractHeadMetadataTwitterCard(t *testing.T) {
	// Twitter Card meta tags occur with both "name" and "property" attributes
	// in the wild. Both should be picked up.
	htmlInput := `<!doctype html>
<html><head>
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="Bluesky post">
<meta property="twitter:description" content="Some description">
<meta name="twitter:image" content="https://cdn.example.org/img.png">
</head><body></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("unable to parse fixture: %v", err)
	}

	metadata := ExtractHeadMetadata(doc)

	if metadata["twitter:card"] != "summary_large_image" {
		t.Errorf("twitter:card not extracted, got %q", metadata["twitter:card"])
	}
	if metadata["twitter:description"] != "Some description" {
		t.Errorf("twitter:description not extracted, got %q", metadata["twitter:description"])
	}
	if metadata["twitter:image"] != "https://cdn.example.org/img.png" {
		t.Errorf("twitter:image not extracted, got %q", metadata["twitter:image"])
	}
}

func TestExtractHeadMetadataIgnoresUnrelatedMeta(t *testing.T) {
	htmlInput := `<!doctype html>
<html><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width">
<meta name="description" content="Plain meta description, not a card">
<meta property="og:title" content="Only this should be kept">
</head><body></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("unable to parse fixture: %v", err)
	}

	metadata := ExtractHeadMetadata(doc)

	if len(metadata) != 1 {
		t.Fatalf("expected exactly 1 entry, got %d (%v)", len(metadata), metadata)
	}
	if metadata["og:title"] != "Only this should be kept" {
		t.Errorf("og:title missing, got map = %v", metadata)
	}
}

func TestExtractHeadMetadataKeepsFirstValue(t *testing.T) {
	htmlInput := `<!doctype html>
<html><head>
<meta property="og:image" content="https://cdn.example.org/first.jpg">
<meta property="og:image" content="https://cdn.example.org/second.jpg">
</head><body></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("unable to parse fixture: %v", err)
	}

	metadata := ExtractHeadMetadata(doc)

	if metadata["og:image"] != "https://cdn.example.org/first.jpg" {
		t.Errorf("expected first og:image to win, got %q", metadata["og:image"])
	}
}

func TestExtractHeadMetadataIgnoresEmptyContent(t *testing.T) {
	htmlInput := `<!doctype html>
<html><head>
<meta property="og:title" content="">
<meta property="og:description" content="   ">
<meta property="og:image" content="https://cdn.example.org/x.jpg">
</head><body></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("unable to parse fixture: %v", err)
	}

	metadata := ExtractHeadMetadata(doc)

	if _, ok := metadata["og:title"]; ok {
		t.Errorf("og:title with empty content should be ignored")
	}
	if _, ok := metadata["og:description"]; ok {
		t.Errorf("og:description with whitespace-only content should be ignored")
	}
	if metadata["og:image"] != "https://cdn.example.org/x.jpg" {
		t.Errorf("og:image not extracted, got %q", metadata["og:image"])
	}
}
