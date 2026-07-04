// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package discord

import (
	"strings"
	"testing"
	"time"

	"miniflux.app/v2/internal/model"
)

func TestRenderTemplateBasic(t *testing.T) {
	data := templateData{
		Title:       "Test Entry",
		URL:         "https://example.com/entry",
		Description: "Hello world",
		Author:      "Author",
		FeedTitle:   "Test Feed",
	}

	result := renderTemplate("**{title}**\n{url}", data)
	expected := "**Test Entry**\nhttps://example.com/entry"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderTemplateAllTokens(t *testing.T) {
	data := templateData{
		Title:       "Title",
		URL:         "https://example.com",
		Description: "Desc",
		Author:      "Author",
		Date:        "2026-07-04T10:00:00Z",
		FeedTitle:   "Feed",
		FeedURL:     "https://feed.example.com",
	}

	tpl := "{title}\n{url}\n{description}\n{author}\n{date}\n{feed_title}\n{feed_url}"
	result := renderTemplate(tpl, data)
	expected := "Title\nhttps://example.com\nDesc\nAuthor\n2026-07-04T10:00:00Z\nFeed\nhttps://feed.example.com"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderTemplateEmpty(t *testing.T) {
	data := templateData{Title: "Test"}
	result := renderTemplate("", data)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestRenderTemplateNoTokens(t *testing.T) {
	data := templateData{Title: "Test"}
	result := renderTemplate("Hello world", data)
	if result != "Hello world" {
		t.Errorf("expected %q, got %q", "Hello world", result)
	}
}

func TestRenderTemplateMissingToken(t *testing.T) {
	data := templateData{Title: "Test"}
	result := renderTemplate("{title} by {author}", data)
	expected := "Test by "
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderTemplateTruncation(t *testing.T) {
	data := templateData{
		Description: string(make([]rune, 4500)),
	}
	result := renderTemplate("{description}", data)
	if len([]rune(result)) > 4000 {
		t.Errorf("expected truncation to 4000, got %d", len([]rune(result)))
	}
}

func TestValidateTemplateValid(t *testing.T) {
	tests := []string{
		"{title}",
		"**{title}**\n{url}",
		"{title}\n{description}\n{url}\n{author}\n{date}\n{feed_title}\n{feed_url}",
		"Hello world",
		"",
	}
	for _, tpl := range tests {
		if !ValidateTemplate(tpl) {
			t.Errorf("expected valid template: %q", tpl)
		}
	}
}

func TestValidateTemplateInvalid(t *testing.T) {
	tests := []string{
		"{nonexistent}",
		"{title} {foo}",
		"test {bar} end",
	}
	for _, tpl := range tests {
		if ValidateTemplate(tpl) {
			t.Errorf("expected invalid template: %q", tpl)
		}
	}
}

func TestValidateTemplatePartialBrace(t *testing.T) {
	tpl := "Hello { world"
	if !ValidateTemplate(tpl) {
		t.Error("expected valid template for unmatched open brace")
	}
}

func TestBuildTemplateData(t *testing.T) {
	feed := &model.Feed{
		Title:   "Test Feed",
		SiteURL: "https://example.com",
	}

	entryDate := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	entry := &model.Entry{
		Title:   "Test Entry",
		URL:     "https://example.com/entry",
		Author:  "Author",
		Content: "<p>Hello <b>world</b></p>",
		Date:    entryDate,
	}

	data := buildTemplateData(feed, entry)

	if data.Title != "Test Entry" {
		t.Errorf("expected Title %q, got %q", "Test Entry", data.Title)
	}
	if data.URL != "https://example.com/entry" {
		t.Errorf("expected URL %q, got %q", "https://example.com/entry", data.URL)
	}
	if data.Author != "Author" {
		t.Errorf("expected Author %q, got %q", "Author", data.Author)
	}
	if data.FeedTitle != "Test Feed" {
		t.Errorf("expected FeedTitle %q, got %q", "Test Feed", data.FeedTitle)
	}
	if data.FeedURL != "https://example.com" {
		t.Errorf("expected FeedURL %q, got %q", "https://example.com", data.FeedURL)
	}
	if data.Date != "2026-07-04T10:00:00Z" {
		t.Errorf("expected Date %q, got %q", "2026-07-04T10:00:00Z", data.Date)
	}
}

func TestBuildTemplateDataHTMLStripping(t *testing.T) {
	feed := &model.Feed{Title: "Feed", SiteURL: "https://example.com"}
	entry := &model.Entry{
		Title:   "Entry",
		URL:     "https://example.com/entry",
		Content: "<p>Hello <b>world</b> &amp; friends</p>",
		Date:    time.Now(),
	}

	data := buildTemplateData(feed, entry)
	if data.Description != "Hello world & friends" {
		t.Errorf("expected %q, got %q", "Hello world & friends", data.Description)
	}
}

func TestBuildTemplateDataContentTruncation(t *testing.T) {
	feed := &model.Feed{Title: "Feed", SiteURL: "https://example.com"}
	entry := &model.Entry{
		Title:   "Entry",
		URL:     "https://example.com/entry",
		Content: string(make([]byte, 600)),
		Date:    time.Now(),
	}

	data := buildTemplateData(feed, entry)
	if !strings.HasSuffix(data.Description, "...") {
		t.Errorf("expected truncation suffix '...', got ending: %q", data.Description[len(data.Description)-10:])
	}
	if len([]rune(data.Description)) != 503 {
		t.Errorf("expected 503 runes (500 + '...'), got %d", len([]rune(data.Description)))
	}
}

func TestBuildTemplateDataShortContent(t *testing.T) {
	feed := &model.Feed{Title: "Feed", SiteURL: "https://example.com"}
	entry := &model.Entry{
		Title:   "Entry",
		URL:     "https://example.com/entry",
		Content: "Short content",
		Date:    time.Now(),
	}

	data := buildTemplateData(feed, entry)
	if strings.HasSuffix(data.Description, "...") {
		t.Error("short content should not have truncation suffix")
	}
	if data.Description != "Short content" {
		t.Errorf("expected %q, got %q", "Short content", data.Description)
	}
}

func TestBuildTemplateDataZeroDate(t *testing.T) {
	feed := &model.Feed{Title: "Feed", SiteURL: "https://example.com"}
	entry := &model.Entry{
		Title: "Entry",
		URL:   "https://example.com/entry",
		Date:  time.Time{},
	}

	data := buildTemplateData(feed, entry)
	if data.Date != "" {
		t.Errorf("expected empty date for zero time, got %q", data.Date)
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<b>Bold</b> and <i>italic</i>", "Bold and italic"},
		{"No tags", "No tags"},
		{"<br/>", ""},
		{"<p class=\"foo\">Bar</p>", "Bar"},
	}

	for _, tt := range tests {
		result := stripHTMLTags(tt.input)
		if result != tt.expected {
			t.Errorf("stripHTMLTags(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
