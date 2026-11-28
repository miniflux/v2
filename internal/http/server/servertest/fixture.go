// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package servertest // import "miniflux.app/v2/internal/http/server/servertest"

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
)

const (
	FeedTitle         = "Integration Test Feed"
	SubscriptionTitle = "Integration Test Subscription"
	EntryCount        = 3

	// SearchTerm appears only in the content of the second entry.
	SearchTerm = "zanzibar"

	// ArticleText appears in every article page, but not in the feed.
	ArticleText = "This paragraph only exists on the original article page."
)

// FeedFixture is a local website with a feed, an icon and article pages.
// The server under test fetches it instead of a remote site.
type FeedFixture struct {
	WebsiteURL string
	FeedURL    string
}

// NewFeedFixture starts a fixture website that is shut down when the test ends.
func NewFeedFixture(t testing.TB) *FeedFixture {
	t.Helper()

	icon := newFixtureIcon(t)
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	fixture := &FeedFixture{
		WebsiteURL: server.URL + "/",
		FeedURL:    server.URL + "/feed.xml",
	}

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<title>Integration Test Website</title>
	<link rel="alternate" type="application/rss+xml" title=%q href="/feed.xml">
	<link rel="icon" type="image/png" href="/icon.png">
</head>
<body><h1>Integration Test Website</h1></body>
</html>`, SubscriptionTitle)
	})

	mux.HandleFunc("GET /feed.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, fixture.feedDocument())
	})

	mux.HandleFunc("GET /icon.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(icon)
	})

	mux.HandleFunc("GET /entries/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Article %[1]s</title></head>
<body>
	<article>
		<h1>Article %[1]s</h1>
		<p>%[2]s It is long enough for the content extractor to keep it as the main content of the page.</p>
		<p>A second paragraph gives the article some more text, so that it does not look like a navigation block.</p>
	</article>
</body>
</html>`, id, ArticleText)
	})

	return fixture
}

// Subscribe creates a feed for the fixture with the given client and returns its ID.
func (f *FeedFixture) Subscribe(t testing.TB, client *miniflux.Client) int64 {
	t.Helper()

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: f.FeedURL})
	if err != nil {
		t.Fatal(err)
	}
	return feedID
}

// FeedEntries returns the entries of a fixture feed oldest first: entry 1 is published on 2025-01-01, entry 3 on 2025-01-03.
func FeedEntries(t testing.TB, client *miniflux.Client, feedID int64) miniflux.Entries {
	t.Helper()

	result, err := client.FeedEntries(feedID, &miniflux.Filter{Order: "published_at", Direction: "asc"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != EntryCount {
		t.Fatalf(`Expected %d entries, got %d`, EntryCount, len(result.Entries))
	}
	return result.Entries
}

// feedDocument returns an RSS feed with EntryCount entries published a day apart.
// The second entry mentions SearchTerm and has an audio enclosure.
func (f *FeedFixture) feedDocument() string {
	var feed strings.Builder
	fmt.Fprintf(&feed, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
	<title>%s</title>
	<link>%s</link>
	<description>Feed served by the integration tests</description>`, FeedTitle, f.WebsiteURL)

	for index := 1; index <= EntryCount; index++ {
		content := fmt.Sprintf("Content of entry %d.", index)
		enclosure := ""
		if index == 2 {
			content += " It mentions " + SearchTerm + "."
			enclosure = fmt.Sprintf(`<enclosure url="%smedia/episode-%d.mp3" length="1024" type="audio/mpeg"/>`, f.WebsiteURL, index)
		}

		fmt.Fprintf(&feed, `
	<item>
		<guid isPermaLink="false">entry-%[1]d</guid>
		<title>Entry %[1]d</title>
		<link>%[2]sentries/%[1]d</link>
		<pubDate>%[3]s</pubDate>
		<description>%[4]s</description>
		%[5]s
	</item>`, index, f.WebsiteURL, time.Date(2025, time.January, index, 10, 0, 0, 0, time.UTC).Format(time.RFC1123Z), content, enclosure)
	}

	feed.WriteString("\n</channel>\n</rss>\n")
	return feed.String()
}

func newFixtureIcon(t testing.TB) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for column := range 16 {
		for row := range 16 {
			img.Set(column, row, color.RGBA{R: 0x33, G: 0x66, B: 0x99, A: 0xff})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
