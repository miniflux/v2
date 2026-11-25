// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestRefreshFeedPreservesLocalEdits(t *testing.T) {
	t.Parallel()

	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	admin := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	user, err := admin.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.DeleteUser(user.ID); err != nil {
			t.Error(err)
		}
	})
	client := miniflux.NewClient(testConfig.testBaseURL, user.Username, testConfig.testRegularPassword)

	originalItem := `<item><guid isPermaLink="false">first</guid><title>First article</title>
		<link>https://example.org/first</link><description><![CDATA[<p>Original content</p>]]></description></item>`
	const newItem = `<item><guid isPermaLink="false">second</guid><title>Second article</title>
		<link>https://example.org/second</link><description>New article</description></item>`
	var publisherItems atomic.Pointer[string]
	publisherItems.Store(&originalItem)
	publisher := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/feed.xml" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		io.WriteString(w, `<rss version="2.0"><channel><title>Publisher feed</title><link>http://`+
			r.Host+`/</link><description>Publisher feed</description>`+*publisherItems.Load()+`</channel></rss>`)
	}))
	t.Cleanup(publisher.Close)

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: publisher.URL + "/feed.xml"})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := client.FeedEntries(feedID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries.Entries) != 1 {
		t.Fatalf("Expected one initial entry, got %d", len(entries.Entries))
	}
	entryID := entries.Entries[0].ID

	assertEntry := func(title, content string) {
		t.Helper()
		entry, err := client.FeedEntry(feedID, entryID)
		if err != nil {
			t.Fatal(err)
		}
		if entry.Title != title || entry.Content != content {
			t.Fatalf("Expected entry title %q and content %q, got %q and %q", title, content, entry.Title, entry.Content)
		}
	}

	localTitle, localContent := "Edited title", "<p>Edited content</p>"
	if _, err := client.UpdateEntry(entryID, &miniflux.EntryModificationRequest{Title: &localTitle, Content: &localContent}); err != nil {
		t.Fatal(err)
	}
	rewriteRules := `replace("Original content"|"Rewritten content")`
	if _, err := client.UpdateFeed(feedID, &miniflux.FeedModificationRequest{RewriteRules: &rewriteRules}); err != nil {
		t.Fatal(err)
	}

	withNewItem := originalItem + newItem
	publisherItems.Store(&withNewItem)
	if err := client.RefreshFeed(feedID); err != nil {
		t.Fatal(err)
	}
	assertEntry(localTitle, localContent)
	entries, err = client.FeedEntries(feedID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries.Entries) != 2 {
		t.Fatalf("Expected both entries after refresh, got %d", len(entries.Entries))
	}

	if err := client.ForceRefreshFeed(feedID); err != nil {
		t.Fatal(err)
	}
	assertEntry("First article", "<p>Rewritten content</p>")

	correctedItem := `<item><guid isPermaLink="false">first</guid><title>Corrected article</title>
		<link>https://example.org/first</link><description><![CDATA[<p>Corrected content</p>]]></description></item>` + newItem
	publisherItems.Store(&correctedItem)
	if err := client.RefreshFeed(feedID); err != nil {
		t.Fatal(err)
	}
	assertEntry("Corrected article", "<p>Corrected content</p>")
}
