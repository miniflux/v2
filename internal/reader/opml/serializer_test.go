// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

import (
	"bytes"
	"strings"
	"testing"
)

func TestSerialize(t *testing.T) {
	var subscriptions []subcription
	subscriptions = append(subscriptions, subcription{Title: "Feed 1", FeedURL: "http://example.org/feed/1", SiteURL: "http://example.org/1", CategoryName: "Category 1"})
	subscriptions = append(subscriptions, subcription{Title: "Feed 2", FeedURL: "http://example.org/feed/2", SiteURL: "http://example.org/2", CategoryName: "Category 1"})
	subscriptions = append(subscriptions, subcription{Title: "Feed 3", FeedURL: "http://example.org/feed/3", SiteURL: "http://example.org/3", CategoryName: "Category 2"})

	output := serialize(subscriptions)
	feeds, err := parse(bytes.NewBufferString(output))
	if err != nil {
		t.Error(err)
	}

	if len(feeds) != 3 {
		t.Errorf("Wrong number of subscriptions: %d instead of %d", len(feeds), 3)
	}

	found := false
	for _, feed := range feeds {
		if feed.Title == "Feed 1" && feed.CategoryName == "Category 1" &&
			feed.FeedURL == "http://example.org/feed/1" && feed.SiteURL == "http://example.org/1" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Serialized feed is incorrect")
	}
}

func TestSerializeWithMinifluxSettings(t *testing.T) {
	input := subcription{
		Title:                       "Feed 1",
		FeedURL:                     "http://example.org/feed/1",
		SiteURL:                     "http://example.org/1",
		CategoryName:                "Category 1",
		ScraperRules:                `article [class^="content"]`,
		RewriteRules:                `replace("foo"|"bar")`,
		UrlRewriteRules:             `rewrite("^https://old"|"https://new")`,
		BlocklistRules:              "sponsored",
		KeeplistRules:               "important",
		BlockFilterEntryRules:       `EntryTitle=~"ad"`,
		KeepFilterEntryRules:        `EntryTitle=~"news"`,
		UserAgent:                   "CustomAgent/1.0",
		ProxyURL:                    "http://proxy.example.org",
		Crawler:                     true,
		IgnoreHTTPCache:             true,
		FetchViaProxy:               true,
		Disabled:                    true,
		NoMediaPlayer:               true,
		HideGlobally:                true,
		AllowSelfSignedCertificates: true,
		DisableHTTP2:                true,
		IgnoreEntryUpdates:          true,
	}

	output := serialize([]subcription{input})
	if !strings.Contains(output, `xmlns:miniflux="https://miniflux.app/opml"`) {
		t.Fatal("Miniflux OPML namespace is missing")
	}

	if !strings.Contains(output, `miniflux:crawler="true"`) {
		t.Fatal("Miniflux settings are not serialized with the Miniflux namespace")
	}

	if !strings.Contains(output, `miniflux:proxyUrl="http://proxy.example.org"`) {
		t.Fatal("Proxy URL is not serialized with the Miniflux namespace")
	}

	if strings.Contains(output, "cookie") {
		t.Fatal("Sensitive feed settings should not be serialized")
	}

	feeds, err := parse(bytes.NewBufferString(output))
	if err != nil {
		t.Fatal(err)
	}

	if len(feeds) != 1 {
		t.Fatalf("Wrong number of subscriptions: %d instead of %d", len(feeds), 1)
	}

	if feeds[0] != input {
		t.Errorf("Round-trip failed:\ngot:  %+v\nwant: %+v", feeds[0], input)
	}
}

func TestSerializePreservesNewlinesInRules(t *testing.T) {
	input := subcription{
		Title:                 "Feed 1",
		FeedURL:               "http://example.org/feed/1",
		SiteURL:               "http://example.org/1",
		CategoryName:          "Category 1",
		RewriteRules:          "replace(\"foo\"|\"bar\")\nadd_youtube_video",
		ScraperRules:          "article.content\np.body",
		BlockFilterEntryRules: "EntryTitle=~\"ad\"\nEntryURL=~\"click\"",
	}

	output := serialize([]subcription{input})
	feeds, err := parse(bytes.NewBufferString(output))
	if err != nil {
		t.Fatal(err)
	}

	if len(feeds) != 1 {
		t.Fatalf("Wrong number of subscriptions: %d instead of %d", len(feeds), 1)
	}

	if feeds[0].RewriteRules != input.RewriteRules {
		t.Errorf("RewriteRules newlines not preserved:\ngot:  %q\nwant: %q", feeds[0].RewriteRules, input.RewriteRules)
	}

	if feeds[0].ScraperRules != input.ScraperRules {
		t.Errorf("ScraperRules newlines not preserved:\ngot:  %q\nwant: %q", feeds[0].ScraperRules, input.ScraperRules)
	}

	if feeds[0].BlockFilterEntryRules != input.BlockFilterEntryRules {
		t.Errorf("BlockFilterEntryRules newlines not preserved:\ngot:  %q\nwant: %q", feeds[0].BlockFilterEntryRules, input.BlockFilterEntryRules)
	}
}

func TestNormalizedCategoriesOrder(t *testing.T) {
	var orderTests = []struct {
		naturalOrderName string
		correctOrderName string
	}{
		{"Category 2", "Category 1"},
		{"Category 3", "Category 2"},
		{"Category 1", "Category 3"},
	}

	var subscriptions []subcription
	subscriptions = append(subscriptions, subcription{Title: "Feed 1", FeedURL: "http://example.org/feed/1", SiteURL: "http://example.org/1", CategoryName: orderTests[0].naturalOrderName})
	subscriptions = append(subscriptions, subcription{Title: "Feed 2", FeedURL: "http://example.org/feed/2", SiteURL: "http://example.org/2", CategoryName: orderTests[1].naturalOrderName})
	subscriptions = append(subscriptions, subcription{Title: "Feed 3", FeedURL: "http://example.org/feed/3", SiteURL: "http://example.org/3", CategoryName: orderTests[2].naturalOrderName})

	feeds := convertSubscriptionsToOPML(subscriptions)

	for i, o := range orderTests {
		if feeds.Outlines[i].Text != o.correctOrderName {
			t.Fatalf("need %v, got %v", o.correctOrderName, feeds.Outlines[i].Text)
		}
	}
}
