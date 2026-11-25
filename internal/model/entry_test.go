// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"testing"
	"time"
)

func TestEntrySourceFingerprintIgnoresNonSourceValues(t *testing.T) {
	entry := &Entry{
		Title:       "Title",
		Content:     "<p>Content</p>",
		URL:         "https://example.org/article",
		CommentsURL: "https://example.org/article#comments",
		Author:      "Author",
		Language:    "en",
		Tags:        []string{"tag1", "tag2"},
		Enclosures: EnclosureList{
			{URL: "https://example.org/audio.mp3", MimeType: "audio/mpeg", Size: 1024},
		},
	}

	expected := entry.ComputeSourceFingerprint()
	if expected == "" {
		t.Fatal("The source fingerprint should not be empty")
	}

	entry.ID = 42
	entry.UserID = 1
	entry.FeedID = 2
	entry.Hash = "identity"
	entry.Status = EntryStatusRead
	entry.Starred = true
	entry.ReadingTime = 12
	entry.Date = time.Now()
	entry.CreatedAt = time.Now()
	entry.ChangedAt = time.Now()
	entry.Enclosures[0].ID = 7
	entry.Enclosures[0].MediaProgression = 512
	entry.Enclosures[0].MimeType = "audio/ogg"
	entry.Enclosures[0].Size = 2048

	if fingerprint := entry.ComputeSourceFingerprint(); fingerprint != expected {
		t.Errorf("The source fingerprint should ignore local values, got %q instead of %q", fingerprint, expected)
	}
}

func TestEntrySourceFingerprintDetectsSourceChanges(t *testing.T) {
	scenarios := []struct {
		name   string
		modify func(*Entry)
	}{
		{"title", func(e *Entry) { e.Title = "New title" }},
		{"content", func(e *Entry) { e.Content = "<p>New content</p>" }},
		{"url", func(e *Entry) { e.URL = "https://example.org/moved" }},
		{"comments_url", func(e *Entry) { e.CommentsURL = "https://example.org/comments" }},
		{"author", func(e *Entry) { e.Author = "Another author" }},
		{"language", func(e *Entry) { e.Language = "fr" }},
		{"tags", func(e *Entry) { e.Tags = []string{"tag2", "tag1"} }},
		{"enclosure_url", func(e *Entry) { e.Enclosures[0].URL = "https://example.org/audio-v2.mp3" }},
		{"enclosure_added", func(e *Entry) {
			e.Enclosures = append(e.Enclosures, &Enclosure{URL: "https://example.org/extra.mp3"})
		}},
		{"enclosure_removed", func(e *Entry) { e.Enclosures = nil }},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			entry := &Entry{
				Title:       "Title",
				Content:     "<p>Content</p>",
				URL:         "https://example.org/article",
				CommentsURL: "https://example.org/article#comments",
				Author:      "Author",
				Language:    "en",
				Tags:        []string{"tag1", "tag2"},
				Enclosures: EnclosureList{
					{URL: "https://example.org/audio.mp3", MimeType: "audio/mpeg", Size: 1024},
				},
			}

			expected := entry.ComputeSourceFingerprint()
			scenario.modify(entry)

			if fingerprint := entry.ComputeSourceFingerprint(); fingerprint == expected {
				t.Errorf("A change of the %s should produce a different source fingerprint", scenario.name)
			}
		})
	}
}

func TestEntrySourceFingerprintIsNotAmbiguous(t *testing.T) {
	first := &Entry{Title: "ab", Content: "c"}
	second := &Entry{Title: "a", Content: "bc"}

	if first.ComputeSourceFingerprint() == second.ComputeSourceFingerprint() {
		t.Error("Moving characters between fields should produce a different source fingerprint")
	}

	withTag := &Entry{Tags: []string{"https://example.org/audio.mp3"}}
	withEnclosure := &Entry{Enclosures: EnclosureList{{URL: "https://example.org/audio.mp3"}}}

	if withTag.ComputeSourceFingerprint() == withEnclosure.ComputeSourceFingerprint() {
		t.Error("Moving a value between tags and enclosures should produce a different source fingerprint")
	}
}

func TestEntrySourceFingerprintUsesUniqueEnclosureURLs(t *testing.T) {
	first := &Entry{Enclosures: EnclosureList{
		{URL: "https://example.org/first.mp3"},
		{URL: "https://example.org/second.mp3"},
	}}
	second := &Entry{Enclosures: EnclosureList{
		{URL: " https://example.org/second.mp3 "},
		{URL: "https://example.org/first.mp3"},
		{URL: "https://example.org/first.mp3"},
	}}

	if first.ComputeSourceFingerprint() != second.ComputeSourceFingerprint() {
		t.Error("The same enclosure URLs should produce the same source fingerprint")
	}
}
