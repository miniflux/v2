// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

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
)

const (
	fixtureFeedTitle         = "Integration Test Feed"
	fixtureSubscriptionTitle = "Integration Test Subscription"
	fixtureEntryCount        = 3

	// fixtureSearchTerm appears only in the content of the second entry.
	fixtureSearchTerm = "zanzibar"

	// fixtureArticleText appears in every article page, but not in the feed.
	fixtureArticleText = "This paragraph only exists on the original article page."
)

// feedFixture is a local website with a feed, an icon and article pages.
// The Miniflux server under test fetches it instead of a remote site,
// which requires FETCHER_ALLOW_PRIVATE_NETWORKS=1.
type feedFixture struct {
	websiteURL string
	feedURL    string
}

// newFeedFixture starts a fixture website that is shut down when the test ends.
func newFeedFixture(t *testing.T) *feedFixture {
	t.Helper()

	icon := newFixtureIcon(t)
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	fixture := &feedFixture{
		websiteURL: server.URL + "/",
		feedURL:    server.URL + "/feed.xml",
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
</html>`, fixtureSubscriptionTitle)
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
</html>`, id, fixtureArticleText)
	})

	return fixture
}

// feedDocument returns an RSS feed with fixtureEntryCount entries published a day apart.
// The second entry mentions fixtureSearchTerm and has an audio enclosure.
func (f *feedFixture) feedDocument() string {
	var feed strings.Builder
	fmt.Fprintf(&feed, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
	<title>%s</title>
	<link>%s</link>
	<description>Feed served by the API integration tests</description>`, fixtureFeedTitle, f.websiteURL)

	for i := 1; i <= fixtureEntryCount; i++ {
		content := fmt.Sprintf("Content of entry %d.", i)
		enclosure := ""
		if i == 2 {
			content += " It mentions " + fixtureSearchTerm + "."
			enclosure = fmt.Sprintf(`<enclosure url="%smedia/episode-%d.mp3" length="1024" type="audio/mpeg"/>`, f.websiteURL, i)
		}

		fmt.Fprintf(&feed, `
	<item>
		<guid isPermaLink="false">entry-%[1]d</guid>
		<title>Entry %[1]d</title>
		<link>%[2]sentries/%[1]d</link>
		<pubDate>%[3]s</pubDate>
		<description>%[4]s</description>
		%[5]s
	</item>`, i, f.websiteURL, time.Date(2025, time.January, i, 10, 0, 0, 0, time.UTC).Format(time.RFC1123Z), content, enclosure)
	}

	feed.WriteString("\n</channel>\n</rss>\n")
	return feed.String()
}

func newFixtureIcon(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for x := range 16 {
		for y := range 16 {
			img.Set(x, y, color.RGBA{R: 0x33, G: 0x66, B: 0x99, A: 0xff})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
