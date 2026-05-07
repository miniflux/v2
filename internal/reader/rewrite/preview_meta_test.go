// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestAddOpenGraphPrependsImageAndDescription(t *testing.T) {
	entry := &model.Entry{
		URL:     "https://example.org/post/1",
		Title:   "Bluesky post",
		Content: "Original body.",
	}
	ctx := &RewriteContext{Metadata: map[string]string{
		"og:description": "I took this picture of a cat today.",
		"og:image":       "https://cdn.example.org/cat.jpg",
	}}

	ApplyContentRewriteRules(entry, `add_open_graph("description","image")`, ctx)

	if !strings.Contains(entry.Content, `src="https://cdn.example.org/cat.jpg"`) {
		t.Errorf("expected og:image to be embedded, got: %s", entry.Content)
	}
	if !strings.Contains(entry.Content, "I took this picture of a cat today.") {
		t.Errorf("expected og:description to be embedded, got: %s", entry.Content)
	}
	if !strings.HasSuffix(entry.Content, "Original body.") {
		t.Errorf("expected original body to remain at the end, got: %s", entry.Content)
	}
}

func TestAddTwitterCardPrependsImageAndDescription(t *testing.T) {
	entry := &model.Entry{
		URL:     "https://example.org/post/1",
		Content: "Original body.",
	}
	ctx := &RewriteContext{Metadata: map[string]string{
		"twitter:description": "Card description",
		"twitter:image":       "https://cdn.example.org/img.png",
	}}

	ApplyContentRewriteRules(entry, `add_twitter_card("description","image")`, ctx)

	if !strings.Contains(entry.Content, `src="https://cdn.example.org/img.png"`) {
		t.Errorf("expected twitter:image to be embedded, got: %s", entry.Content)
	}
	if !strings.Contains(entry.Content, "Card description") {
		t.Errorf("expected twitter:description to be embedded, got: %s", entry.Content)
	}
}

func TestAddOpenGraphFallbackToDefaults(t *testing.T) {
	entry := &model.Entry{
		URL:     "https://example.org/post/1",
		Content: "body",
	}
	ctx := &RewriteContext{Metadata: map[string]string{
		"og:description": "default description",
		"og:image":       "https://cdn.example.org/default.jpg",
		"og:title":       "ignored without explicit arg",
	}}

	ApplyContentRewriteRules(entry, `add_open_graph`, ctx)

	if !strings.Contains(entry.Content, "default description") {
		t.Errorf("expected default og:description, got: %s", entry.Content)
	}
	if !strings.Contains(entry.Content, "https://cdn.example.org/default.jpg") {
		t.Errorf("expected default og:image, got: %s", entry.Content)
	}
	if strings.Contains(entry.Content, "ignored without explicit arg") {
		t.Errorf("default arg list should not include title, got: %s", entry.Content)
	}
}

func TestAddOpenGraphAcceptsFullyQualifiedKey(t *testing.T) {
	entry := &model.Entry{URL: "https://example.org/", Content: "body"}
	ctx := &RewriteContext{Metadata: map[string]string{
		"og:description": "qualified key works too",
	}}

	ApplyContentRewriteRules(entry, `add_open_graph("og:description")`, ctx)

	if !strings.Contains(entry.Content, "qualified key works too") {
		t.Errorf("expected fully qualified og:description to work, got: %s", entry.Content)
	}
}

func TestAddOpenGraphIgnoresMismatchedFamilyKey(t *testing.T) {
	entry := &model.Entry{URL: "https://example.org/", Content: "body"}
	ctx := &RewriteContext{Metadata: map[string]string{
		"twitter:description": "from twitter",
	}}

	// Asking add_open_graph for a "twitter:" key should be ignored, not
	// silently fetched from the twitter family.
	ApplyContentRewriteRules(entry, `add_open_graph("twitter:description")`, ctx)

	if strings.Contains(entry.Content, "from twitter") {
		t.Errorf("add_open_graph should not pick twitter:* values, got: %s", entry.Content)
	}
}

func TestAddOpenGraphNoMetadataIsNoop(t *testing.T) {
	entry := &model.Entry{URL: "https://example.org/", Content: "untouched"}

	// No context at all, simulating crawler-disabled feeds.
	ApplyContentRewriteRules(entry, `add_open_graph("description","image")`, nil)

	if entry.Content != "untouched" {
		t.Errorf("expected no-op when no metadata is available, got: %s", entry.Content)
	}
}

func TestAddOpenGraphMissingPropertyIsNoop(t *testing.T) {
	entry := &model.Entry{URL: "https://example.org/", Content: "untouched"}
	ctx := &RewriteContext{Metadata: map[string]string{
		"og:url": "https://example.org/canonical",
	}}

	ApplyContentRewriteRules(entry, `add_open_graph("description","image")`, ctx)

	if entry.Content != "untouched" {
		t.Errorf("expected no-op when requested properties are missing, got: %s", entry.Content)
	}
}

func TestAddOpenGraphRendersExtraPropertyAsParagraph(t *testing.T) {
	entry := &model.Entry{URL: "https://example.org/", Content: "body"}
	ctx := &RewriteContext{Metadata: map[string]string{
		"og:site_name": "Example",
	}}

	ApplyContentRewriteRules(entry, `add_open_graph("site_name")`, ctx)

	if !strings.Contains(entry.Content, "site_name") {
		t.Errorf("expected site_name label, got: %s", entry.Content)
	}
	if !strings.Contains(entry.Content, "Example") {
		t.Errorf("expected site_name value, got: %s", entry.Content)
	}
}

func TestAddOpenGraphEscapesValues(t *testing.T) {
	entry := &model.Entry{URL: "https://example.org/", Content: "body"}
	ctx := &RewriteContext{Metadata: map[string]string{
		"og:description": `<script>alert(1)</script>`,
		"og:image":       `https://cdn.example.org/x.jpg"><script>x</script>`,
	}}

	ApplyContentRewriteRules(entry, `add_open_graph("description","image")`, ctx)

	if strings.Contains(entry.Content, "<script>") {
		t.Errorf("expected script tag in metadata to be escaped, got: %s", entry.Content)
	}
	if !strings.Contains(entry.Content, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Errorf("expected escaped description, got: %s", entry.Content)
	}
}
