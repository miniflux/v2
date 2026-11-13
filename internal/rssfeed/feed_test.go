// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssfeed // import "miniflux.app/v2/internal/rssfeed"

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

func TestMain(m *testing.M) {
	config.Opts = config.NewConfigOptions()
	m.Run()
}

func newTestMetadata() feedMetadata {
	return feedMetadata{
		Title:       `Miniflux & "friends"`,
		Description: "Unread entries",
		Link:        "https://miniflux.example.org/",
		ID:          "https://miniflux.example.org/rss/unread",
		SelfURL:     "https://miniflux.example.org/rss/unread",
		IconURL:     "https://miniflux.example.org/feed-icon/abc",
		Language:    "en",
		Updated:     time.Date(2026, time.November, 13, 10, 0, 0, 0, time.UTC),
		Entries: model.Entries{
			{
				ID:          123,
				Title:       `Entry <script> & "quotes"`,
				URL:         "https://example.org/post?a=1&b=2",
				Author:      "Jane Doe",
				Content:     "<p>Hello <strong>world</strong></p>",
				CommentsURL: "https://example.org/post#comments",
				Date:        time.Date(2026, time.November, 12, 8, 30, 0, 0, time.UTC),
				ChangedAt:   time.Date(2026, time.November, 12, 9, 0, 0, 0, time.UTC),
				Enclosures: model.EnclosureList{
					{URL: "https://example.org/audio.mp3", MimeType: "audio/mpeg", Size: 1024},
				},
				Feed: &model.Feed{
					Title:    "Example Feed",
					Language: "en",
					Category: &model.Category{ID: 7, Title: "Tech"},
					Icon:     &model.FeedIcon{ExternalIconID: "abc"},
				},
			},
		},
	}
}

func assertWellFormedXML(t *testing.T, payload []byte) {
	t.Helper()

	decoder := xml.NewDecoder(bytes.NewReader(payload))
	for {
		if _, err := decoder.Token(); err != nil {
			if err == io.EOF {
				return
			}
			t.Fatalf("invalid XML document: %v\n%s", err, payload)
		}
	}
}

func TestRenderRSS(t *testing.T) {
	payload, err := renderRSS(newTestMetadata())
	if err != nil {
		t.Fatalf("renderRSS returned an error: %v", err)
	}

	assertWellFormedXML(t, payload)

	document := string(payload)
	for _, expected := range []string{
		`<rss version="2.0"`,
		`xmlns:content="http://purl.org/rss/1.0/modules/content/"`,
		`xmlns:dc="http://purl.org/dc/elements/1.1/"`,
		`<content:encoded>`,
		`<atom:link href="https://miniflux.example.org/rss/unread" rel="self" type="application/rss+xml"></atom:link>`,
		`<guid isPermaLink="false">urn:miniflux:entry:123</guid>`,
		`<enclosure url="https://example.org/audio.mp3" type="audio/mpeg" length="1024"></enclosure>`,
		`<dc:creator>Jane Doe</dc:creator>`,
		`<category>Example Feed</category>`,
		`<category>Tech</category>`,
		`<pubDate>Thu, 12 Nov 2026 08:30:00 +0000</pubDate>`,
	} {
		if !strings.Contains(document, expected) {
			t.Errorf("RSS document does not contain %q:\n%s", expected, document)
		}
	}

	// Markup must be escaped, not embedded as raw XML.
	if strings.Contains(document, "<strong>world</strong>") {
		t.Errorf("RSS content should be escaped:\n%s", document)
	}
	if !strings.Contains(document, "&lt;strong&gt;world&lt;/strong&gt;") {
		t.Errorf("RSS content does not contain escaped HTML:\n%s", document)
	}
	if strings.Contains(document, "<script>") {
		t.Errorf("RSS title should be escaped:\n%s", document)
	}
}

func TestRenderAtom(t *testing.T) {
	payload, err := renderAtom(newTestMetadata())
	if err != nil {
		t.Fatalf("renderAtom returned an error: %v", err)
	}

	assertWellFormedXML(t, payload)

	document := string(payload)
	for _, expected := range []string{
		`<feed xmlns="http://www.w3.org/2005/Atom">`,
		`<id>https://miniflux.example.org/rss/unread</id>`,
		`<id>urn:miniflux:entry:123</id>`,
		`<published>2026-11-12T08:30:00Z</published>`,
		`<updated>2026-11-12T09:00:00Z</updated>`,
		`<link href="https://example.org/audio.mp3" rel="enclosure" type="audio/mpeg" length="1024"></link>`,
		`<icon>https://miniflux.example.org/feed-icon/abc</icon>`,
		`<category term="Tech"></category>`,
	} {
		if !strings.Contains(document, expected) {
			t.Errorf("Atom document does not contain %q:\n%s", expected, document)
		}
	}

	if !strings.Contains(document, "&lt;p&gt;Hello &lt;strong&gt;world&lt;/strong&gt;&lt;/p&gt;") {
		t.Errorf("Atom content does not contain escaped HTML:\n%s", document)
	}
}

func TestNewHandlerRegistersRoutes(t *testing.T) {
	handler := NewHandler(nil)

	for _, path := range []string{"/rss/unread", "/atom/unread", "/rss/feed/1", "/atom/search"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

		// Authenticated routes reject anonymous requests instead of returning 404.
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for %s, got %d", path, recorder.Code)
		}
	}

	for _, path := range []string{"/rss/", "/atom/"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for the help page %s, got %d", path, recorder.Code)
		}
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/rss/unknown", nil))
	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected 404 for an unknown route, got %d", recorder.Code)
	}
}

func TestExtractCredentials(t *testing.T) {
	scenarios := []struct {
		name         string
		request      *http.Request
		expectedKey  string
		expectedUser string
	}{
		{
			name:        "query parameter",
			request:     httptest.NewRequest(http.MethodGet, "/rss/unread?api_key=secret", nil),
			expectedKey: "secret",
		},
		{
			name:         "basic authentication",
			request:      httptest.NewRequest(http.MethodGet, "/rss/unread", nil),
			expectedKey:  "secret",
			expectedUser: "jane",
		},
		{
			name: "bearer token",
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/rss/unread", nil)
				r.Header.Set("Authorization", "Bearer secret")
				return r
			}(),
			expectedKey: "secret",
		},
		{
			name:    "no credentials",
			request: httptest.NewRequest(http.MethodGet, "/rss/unread", nil),
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if scenario.expectedUser != "" {
				scenario.request.SetBasicAuth(scenario.expectedUser, scenario.expectedKey)
			}

			key, username := extractCredentials(scenario.request)
			if key != scenario.expectedKey {
				t.Errorf("expected API key %q, got %q", scenario.expectedKey, key)
			}
			if username != scenario.expectedUser {
				t.Errorf("expected username %q, got %q", scenario.expectedUser, username)
			}
		})
	}
}

func TestSelfURLRemovesAPIKey(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/rss/unread?api_key=secret&limit=10", nil)
	got := selfURL(r)
	if strings.Contains(got, "secret") {
		t.Errorf("self URL must not contain the API key: %s", got)
	}
	if !strings.Contains(got, "limit=10") {
		t.Errorf("self URL should preserve other query parameters: %s", got)
	}
}

func TestFeedIDIgnoresQueryParameters(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/atom/unread?api_key=secret&limit=10&mark_read=0", nil)
	got := feedID(r)
	if strings.Contains(got, "?") {
		t.Errorf("feed ID must not contain a query string: %s", got)
	}
	if !strings.HasSuffix(got, "/atom/unread") {
		t.Errorf("feed ID should be built from the request path: %s", got)
	}
}

func TestEntryTitleFallbacks(t *testing.T) {
	scenarios := []struct {
		name     string
		entry    model.Entry
		expected string
	}{
		{"title", model.Entry{Title: "Hello"}, "Hello"},
		{"blank title uses the URL", model.Entry{Title: "   ", URL: "https://example.org/post"}, "https://example.org/post"},
		{"no title and no URL", model.Entry{}, "(untitled)"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if got := entryTitle(&scenario.entry); got != scenario.expected {
				t.Errorf("entryTitle() = %q, want %q", got, scenario.expected)
			}
		})
	}
}

func TestEntryPublishedAtFallbacks(t *testing.T) {
	published := time.Date(2026, time.November, 12, 8, 0, 0, 0, time.UTC)
	created := time.Date(2026, time.November, 10, 8, 0, 0, 0, time.UTC)

	if got := entryPublishedAt(&model.Entry{Date: published, CreatedAt: created}); !got.Equal(published) {
		t.Errorf("expected the publication date, got %v", got)
	}
	if got := entryPublishedAt(&model.Entry{CreatedAt: created}); !got.Equal(created) {
		t.Errorf("expected the creation date as fallback, got %v", got)
	}

	before := time.Now().Add(-time.Minute)
	if got := entryPublishedAt(&model.Entry{}); got.Before(before) {
		t.Errorf("expected a current time as last resort, got %v", got)
	}
}

func TestEntryCategories(t *testing.T) {
	if got := entryCategories(&model.Entry{}); len(got) != 0 {
		t.Errorf("expected no categories without a feed, got %v", got)
	}

	entry := &model.Entry{Feed: &model.Feed{Title: "Example Feed"}}
	if got := entryCategories(entry); len(got) != 1 || got[0] != "Example Feed" {
		t.Errorf("expected only the feed title, got %v", got)
	}

	entry.Feed.Category = &model.Category{Title: "Tech"}
	if got := entryCategories(entry); len(got) != 2 || got[1] != "Tech" {
		t.Errorf("expected the feed title and category, got %v", got)
	}
}

func TestFeedIconURL(t *testing.T) {
	if got := feedIconURL(nil); got != "" {
		t.Errorf("expected an empty icon URL for a nil feed, got %q", got)
	}
	if got := feedIconURL(&model.Feed{}); got != "" {
		t.Errorf("expected an empty icon URL without an icon, got %q", got)
	}
	if got := feedIconURL(&model.Feed{Icon: &model.FeedIcon{}}); got != "" {
		t.Errorf("expected an empty icon URL without an external icon ID, got %q", got)
	}

	got := feedIconURL(&model.Feed{Icon: &model.FeedIcon{ExternalIconID: "abc"}})
	if !strings.HasSuffix(got, "/feed-icon/abc") {
		t.Errorf("unexpected icon URL: %q", got)
	}
}

func TestRenderEntriesWithoutFeedMetadata(t *testing.T) {
	entries := model.Entries{
		&model.Entry{ID: 1, URL: "https://example.org/post"},
	}

	documents := map[string]func(feedMetadata) ([]byte, error){
		"rss":  renderRSS,
		"atom": renderAtom,
	}

	for name, render := range documents {
		t.Run(name, func(t *testing.T) {
			payload, err := render(feedMetadata{
				Title:   "Unread",
				Updated: time.Now(),
				Entries: entries,
			})
			if err != nil {
				t.Fatalf("rendering failed: %v", err)
			}

			assertWellFormedXML(t, payload)
			if !strings.Contains(string(payload), "https://example.org/post") {
				t.Errorf("expected the entry title fallback to be used:\n%s", payload)
			}
		})
	}
}
