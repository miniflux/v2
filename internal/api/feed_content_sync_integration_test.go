// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
)

func TestFeedContentChangesIncrementalSync(t *testing.T) {
	t.Parallel()
	config := newIntegrationTestConfig()
	if !config.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	admin := miniflux.NewClient(config.testBaseURL, config.testAdminUsername, config.testAdminPassword)
	user, err := admin.CreateUser(config.genRandomUsername(), config.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.DeleteUser(user.ID) })
	client := miniflux.NewClient(config.testBaseURL, user.Username, config.testRegularPassword)

	title, content, author := "Original title", "Original body", "Original author"
	link, comments := "https://example.org/article", "https://example.org/comments"
	language, tags, enclosures := "en", "", ""
	website := ""
	var document atomic.Value
	publish := func() {
		document.Store(fmt.Sprintf(`<?xml version="1.0"?>
<rss version="2.0"><channel><title>Sync test</title><link>%s</link>
<description>Sync test</description><language>%s</language>
<item><guid isPermaLink="false">stable-entry</guid><title>%s</title><link>%s</link>
<description>%s</description><author>%s</author><comments>%s</comments>
<pubDate>Mon, 01 Sep 2025 12:00:00 GMT</pubDate>%s%s</item>
</channel></rss>`, website, language, title, link, content, author, comments, tags, enclosures))
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprint(w, document.Load().(string))
	}))
	defer server.Close()
	website = server.URL
	publish()
	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.FeedEntries(feedID, nil)
	if err != nil || result.Total != 1 {
		t.Fatalf("Expected one initial entry: result=%+v error=%v", result, err)
	}
	entryID := result.Entries[0].ID
	if err := client.UpdateEntries([]int64{entryID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateEntriesStarred([]int64{entryID}, true); err != nil {
		t.Fatal(err)
	}
	previous, err := client.Entry(entryID)
	if err != nil {
		t.Fatal(err)
	}
	// The API's changed_after cursor has second precision. Cross its boundary
	// once so the unchanged refresh cannot match the initial read/star updates.
	cursor := previous.ChangedAt.Truncate(time.Second).Add(time.Second)
	time.Sleep(time.Until(cursor))

	const audio1 = `<enclosure url="https://example.org/one.mp3" type="audio/mpeg" length="100"/>`
	const audio2 = `<enclosure url="https://example.org/two.mp3" type="audio/mpeg" length="200"/>`
	tests := []struct {
		name    string
		change  func()
		changed bool
	}{
		{"unchanged refresh", func() {}, false},
		{"title", func() { title = "Revised title" }, true},
		{"content", func() { content = "Revised body" }, true},
		{"author", func() { author = "Revised author" }, true},
		{"url", func() { link = "https://example.org/revised" }, true},
		{"comments", func() { comments = "https://example.org/revised-comments" }, true},
		{"tags", func() { tags = "<category>news</category>" }, true},
		{"language", func() { language = "fr" }, true},
		{"add enclosures", func() { enclosures = audio1 + audio2 }, true},
		{"reorder and duplicate enclosures", func() { enclosures = audio2 + audio1 + audio1 }, false},
		{"remove one enclosure", func() { enclosures = audio1 }, true},
		{"remove all enclosures", func() { enclosures = "" }, true},
	}
	for _, test := range tests {
		if !t.Run(test.name, func(t *testing.T) {
			test.change()
			publish()
			if err := client.RefreshFeed(feedID); err != nil {
				t.Fatal(err)
			}
			current, err := client.Entry(entryID)
			if err != nil {
				t.Fatal(err)
			}
			if test.changed && !current.ChangedAt.After(previous.ChangedAt) {
				t.Fatalf("Content change was not tracked: before=%s after=%s", previous.ChangedAt, current.ChangedAt)
			}
			if !test.changed && !current.ChangedAt.Equal(previous.ChangedAt) {
				t.Fatalf("Unchanged content advanced changed_at: before=%s after=%s", previous.ChangedAt, current.ChangedAt)
			}
			if current.Status != miniflux.EntryStatusRead || !current.Starred || !current.Date.Equal(previous.Date) {
				t.Fatal("Refreshing content changed read/starred state or publication date")
			}
			if current.Title != title || current.Content != content {
				t.Fatalf("Feed changes were not persisted: %+v", current)
			}
			updated, err := client.Entries(&miniflux.Filter{ChangedAfter: cursor.Unix()})
			if err != nil {
				t.Fatal(err)
			}
			if test.name == "unchanged refresh" {
				if updated.Total != 0 {
					t.Fatalf("Unchanged entry appeared in incremental sync: %+v", updated)
				}
			} else if updated.Total != 1 || updated.Entries[0].ID != entryID {
				t.Fatalf("Revised entry missing from incremental sync: %+v", updated)
			}
			// An identical refresh must not repeatedly invalidate client caches.
			if err := client.RefreshFeed(feedID); err != nil {
				t.Fatal(err)
			}
			repeated, err := client.Entry(entryID)
			if err != nil {
				t.Fatal(err)
			}
			if !repeated.ChangedAt.Equal(current.ChangedAt) {
				t.Fatal("Repeated refresh advanced changed_at")
			}
			previous = current
		}) {
			return
		}
	}
}
