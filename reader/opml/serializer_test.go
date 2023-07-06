// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/reader/opml"

import (
	"bytes"
	"testing"
)

func TestSerialize(t *testing.T) {
	var subscriptions SubcriptionList
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 1", FeedURL: "http://example.org/feed/1", SiteURL: "http://example.org/1", CategoryName: "Category 1"})
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 2", FeedURL: "http://example.org/feed/2", SiteURL: "http://example.org/2", CategoryName: "Category 1"})
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 3", FeedURL: "http://example.org/feed/3", SiteURL: "http://example.org/3", CategoryName: "Category 2"})

	output := Serialize(subscriptions)
	feeds, err := Parse(bytes.NewBufferString(output))
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

func TestNormalizedCategoriesOrder(t *testing.T) {
	var orderTests = []struct {
		naturalOrderName string
		correctOrderName string
	}{
		{"Category 2", "Category 1"},
		{"Category 3", "Category 2"},
		{"Category 1", "Category 3"},
	}

	var subscriptions SubcriptionList
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 1", FeedURL: "http://example.org/feed/1", SiteURL: "http://example.org/1", CategoryName: orderTests[0].naturalOrderName})
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 2", FeedURL: "http://example.org/feed/2", SiteURL: "http://example.org/2", CategoryName: orderTests[1].naturalOrderName})
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 3", FeedURL: "http://example.org/feed/3", SiteURL: "http://example.org/3", CategoryName: orderTests[2].naturalOrderName})

	feeds := convertSubscriptionsToOPML(subscriptions)

	for i, o := range orderTests {
		if feeds.Outlines[i].Text != o.correctOrderName {
			t.Fatalf("need %v, got %v", o.correctOrderName, feeds.Outlines[i].Text)
		}
	}
}
