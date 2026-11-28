// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader_test

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

// newOwnerWithPrivateFeed creates a user with a feed in a "Private" category and returns the user,
// the feed ID and its entries.
func newOwnerWithPrivateFeed(t *testing.T) (*testUser, int64, miniflux.Entries) {
	t.Helper()

	owner := newTestUser(t)
	category := createCategory(t, owner, "Private")
	feedID, err := owner.rest.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    servertest.NewFeedFixture(t).FeedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	return owner, feedID, servertest.FeedEntries(t, owner.rest, feedID)
}

func TestUsersCannotReadOtherUsersData(t *testing.T) {
	t.Parallel()

	_, feedID, entries := newOwnerWithPrivateFeed(t)
	intruder := newTestUser(t)

	streams := []struct {
		name   string
		stream string
	}{
		{"item IDs of the reading list", readingListStream},
		{"item IDs of the feed", feedStream(feedID)},
	}

	for _, scenario := range streams {
		t.Run(scenario.name, func(t *testing.T) {
			var result itemIDsResult
			intruder.reader.getJSON(t, "/reader/api/0/stream/items/ids", url.Values{"s": {scenario.stream}}, &result)
			if len(result.ItemRefs) != 0 {
				t.Errorf(`Expected no item, got %v`, result.ids())
			}
		})
	}

	t.Run("item contents", func(t *testing.T) {
		var result itemContentsResult
		intruder.reader.postJSON(t, "/reader/api/0/stream/items/contents", url.Values{"i": decimalItemIDs(entries)}, &result)
		if len(result.Items) != 0 {
			t.Errorf(`Expected no item, got %+v`, result.Items)
		}
	})

	t.Run("subscription list", func(t *testing.T) {
		var result subscriptionListResult
		intruder.reader.getJSON(t, "/reader/api/0/subscription/list", nil, &result)
		if len(result.Subscriptions) != 0 {
			t.Errorf(`Expected no subscription, got %+v`, result.Subscriptions)
		}
	})

	t.Run("tag list", func(t *testing.T) {
		var result tagListResult
		intruder.reader.getJSON(t, "/reader/api/0/tag/list", nil, &result)

		expected := []tagResult{
			{ID: fmt.Sprintf("user/%d/state/com.google/starred", intruder.id)},
			{ID: fmt.Sprintf("user/%d/label/All", intruder.id), Label: "All", Type: "folder"},
		}
		if !slices.Equal(result.Tags, expected) {
			t.Errorf(`Expected only the tags of the intruder, got %+v`, result.Tags)
		}
	})
}

func TestUsersCannotChangeOtherUsersData(t *testing.T) {
	t.Parallel()

	owner, feedID, entries := newOwnerWithPrivateFeed(t)
	intruder := newTestUser(t)

	// The requests either succeed without changing anything, or fail as if the data did not exist.
	scenarios := []struct {
		name           string
		path           string
		form           url.Values
		expectedStatus int
	}{
		{"mark the items as read and starred", "/reader/api/0/edit-tag", url.Values{"i": decimalItemIDs(entries), "a": {readStream, starredStream}}, http.StatusOK},
		{"mark the feed as read", "/reader/api/0/mark-all-as-read", url.Values{"s": {feedStream(feedID)}}, http.StatusNotFound},
		{"mark the label as read", "/reader/api/0/mark-all-as-read", url.Values{"s": {labelStream("Private")}}, http.StatusNotFound},
		{"rename the feed", "/reader/api/0/subscription/edit", url.Values{"ac": {"edit"}, "s": {feedStream(feedID)}, "t": {"Stolen"}}, http.StatusBadRequest},
		{"move the feed", "/reader/api/0/subscription/edit", url.Values{"ac": {"edit"}, "s": {feedStream(feedID)}, "a": {labelStream("Stolen")}}, http.StatusBadRequest},
		{"unsubscribe from the feed", "/reader/api/0/subscription/edit", url.Values{"ac": {"unsubscribe"}, "s": {feedStream(feedID)}}, http.StatusOK},
		{"rename the label", "/reader/api/0/rename-tag", url.Values{"s": {labelStream("Private")}, "dest": {labelStream("Stolen")}}, http.StatusNotFound},
		{"remove the label", "/reader/api/0/disable-tag", url.Values{"s": {labelStream("Private")}}, http.StatusOK},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			status, body := intruder.reader.post(t, scenario.path, scenario.form)
			requireStatus(t, status, body, scenario.expectedStatus)
		})
	}

	feed, err := owner.rest.Feed(feedID)
	if err != nil {
		t.Fatalf(`Expected the feed of the owner to remain: %v`, err)
	}
	if feed.Title != servertest.FeedTitle || feed.Category.Title != "Private" {
		t.Errorf(`Expected the feed to be unchanged, got %q in %q`, feed.Title, feed.Category.Title)
	}

	for _, entry := range servertest.FeedEntries(t, owner.rest, feedID) {
		if entry.Status != miniflux.EntryStatusUnread || entry.Starred {
			t.Errorf(`Expected entry %d to be unchanged, got %s with starred=%v`, entry.ID, entry.Status, entry.Starred)
		}
	}

	if titles := categoryTitles(t, owner); !slices.Equal(titles, []string{"All", "Private"}) {
		t.Errorf(`Expected the categories of the owner to be unchanged, got %v`, titles)
	}
	if titles := categoryTitles(t, intruder); !slices.Equal(titles, []string{"All"}) {
		t.Errorf(`Expected the categories of the intruder to be unchanged, got %v`, titles)
	}
}
