// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssfeed

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"miniflux.app/v2/internal/model"
)

// benchmarkContent returns an article body similar in size and shape to what
// most feeds publish (markup with links and media to rewrite).
func benchmarkContent() string {
	paragraph := `<p>Miniflux is a minimalist and opinionated feed reader. It supports ` +
		`<a href="https://example.org/features">many features</a> and can rewrite ` +
		`<img src="http://cdn.example.org/logo.png" alt="logo"> media through its proxy.</p>`
	return strings.Repeat(paragraph, 40)
}

func benchmarkMetadata(entryCount int) feedMetadata {
	content := benchmarkContent()
	feed := &model.Feed{
		ID:       1,
		Title:    "Example Feed",
		SiteURL:  "https://example.org",
		Language: "en",
		Category: &model.Category{ID: 1, Title: "Tech"},
	}
	published := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	entries := make(model.Entries, 0, entryCount)
	for i := range entryCount {
		entries = append(entries, &model.Entry{
			ID:        int64(i + 1),
			FeedID:    feed.ID,
			Title:     fmt.Sprintf("Entry %d", i+1),
			URL:       fmt.Sprintf("https://example.org/posts/%d", i+1),
			Content:   content,
			Author:    "Jane Doe",
			Date:      published.Add(time.Duration(i) * time.Minute),
			ChangedAt: published.Add(time.Duration(i) * time.Minute),
			CreatedAt: published.Add(time.Duration(i) * time.Minute),
			Enclosures: []*model.Enclosure{
				{URL: "https://example.org/audio.mp3", MimeType: "audio/mpeg", Size: 1024},
			},
			Feed: feed,
		})
	}

	return feedMetadata{
		Title:       "Unread",
		Description: "Unread entries from Miniflux",
		Link:        "https://example.org/unread",
		ID:          "https://example.org/rss/unread",
		SelfURL:     "https://example.org/rss/unread",
		Updated:     time.Now(),
		Entries:     entries,
	}
}

func BenchmarkRenderFeed(b *testing.B) {
	formats := map[string]func(feedMetadata) ([]byte, error){
		"rss":  renderRSS,
		"atom": renderAtom,
	}

	for _, entryCount := range []int{50, 200, 1000} {
		meta := benchmarkMetadata(entryCount)

		for name, render := range formats {
			b.Run(fmt.Sprintf("%s/%d_entries", name, entryCount), func(b *testing.B) {
				b.ReportAllocs()

				var payloadSize int
				for b.Loop() {
					payload, err := render(meta)
					if err != nil {
						b.Fatalf("rendering failed: %v", err)
					}
					payloadSize = len(payload)
				}

				b.ReportMetric(float64(payloadSize), "payload-bytes")
			})
		}
	}
}
