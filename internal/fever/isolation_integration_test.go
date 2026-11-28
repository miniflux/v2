// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

// newOwnerWithPrivateFeed creates a user with a feed in a "Private" category and its second entry starred.
// It returns the user, the category, the feed ID and the entries.
func newOwnerWithPrivateFeed(t *testing.T) (*testUser, *miniflux.Category, int64, miniflux.Entries) {
	t.Helper()

	owner := newTestUser(t)
	category := createCategory(t, owner, "Private")
	feedID := subscribeInCategory(t, owner, category.ID)
	entries := servertest.FeedEntries(t, owner.rest, feedID)
	if err := owner.rest.ToggleStarred(entries[1].ID); err != nil {
		t.Fatal(err)
	}
	return owner, category, feedID, entries
}

func TestUsersCannotReadOtherUsersData(t *testing.T) {
	t.Parallel()

	_, _, _, entries := newOwnerWithPrivateFeed(t)

	// The intruder has no feeds: the fixture icon is shared, so their own feed would bring the same favicon.
	intruder := newTestUser(t)

	itemScenarios := []struct {
		name   string
		params url.Values
	}{
		{"oldest items", nil},
		{"most recent items", url.Values{"max_id": {"0"}}},
		{"items with the IDs of the other user", url.Values{"with_ids": {strings.Join(formatIDs(entryIDs(entries)), ",")}}},
	}

	for _, scenario := range itemScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			query := selector("items")
			for key, values := range scenario.params {
				query[key] = values
			}

			var result itemsResult
			intruder.fever.getJSON(t, query, &result)
			if len(result.Items) != 0 || result.TotalItems != 0 {
				t.Errorf(`Expected no item, got %d of %d: %v`, len(result.Items), result.TotalItems, result.ids())
			}
		})
	}

	t.Run("unread item IDs", func(t *testing.T) {
		var result unreadItemIDsResult
		intruder.fever.getJSON(t, selector("unread_item_ids"), &result)
		if result.UnreadItemIDs != "" {
			t.Errorf(`Expected no unread item, got %q`, result.UnreadItemIDs)
		}
	})

	t.Run("saved item IDs", func(t *testing.T) {
		var result savedItemIDsResult
		intruder.fever.getJSON(t, selector("saved_item_ids"), &result)
		if result.SavedItemIDs != "" {
			t.Errorf(`Expected no saved item, got %q`, result.SavedItemIDs)
		}
	})

	t.Run("feeds", func(t *testing.T) {
		var result feedsResult
		intruder.fever.getJSON(t, selector("feeds"), &result)
		if len(result.Feeds) != 0 || len(result.FeedsGroups) != 0 {
			t.Errorf(`Expected no feed, got %+v and %+v`, result.Feeds, result.FeedsGroups)
		}
	})

	t.Run("groups", func(t *testing.T) {
		categories, err := intruder.rest.Categories()
		if err != nil {
			t.Fatal(err)
		}
		expected := make([]groupResult, 0, len(categories))
		for _, category := range categories {
			expected = append(expected, groupResult{ID: category.ID, Title: category.Title})
		}

		var result groupsResult
		intruder.fever.getJSON(t, selector("groups"), &result)
		requireGroups(t, result.Groups, expected)
		if len(result.FeedsGroups) != 0 {
			t.Errorf(`Expected no feeds_groups, got %+v`, result.FeedsGroups)
		}
	})

	t.Run("favicons", func(t *testing.T) {
		var result faviconsResult
		intruder.fever.getJSON(t, selector("favicons"), &result)
		if len(result.Favicons) != 0 {
			t.Errorf(`Expected no favicon, got %+v`, result.Favicons)
		}
	})
}

func TestUsersCannotChangeOtherUsersData(t *testing.T) {
	t.Parallel()

	owner, category, feedID, entries := newOwnerWithPrivateFeed(t)
	intruder := newTestUser(t)
	now := time.Now().Unix()

	// Fever does not report errors: every request gets the authenticated response and changes nothing.
	scenarios := []struct {
		name string
		send func(t *testing.T)
	}{
		{"mark an item as read", func(t *testing.T) { markItem(t, intruder.fever.postJSON, entries[0].ID, "read") }},
		{"mark an item as saved", func(t *testing.T) { markItem(t, intruder.fever.postJSON, entries[0].ID, "saved") }},
		{"mark a saved item as unsaved", func(t *testing.T) { markItem(t, intruder.fever.postJSON, entries[1].ID, "unsaved") }},
		{"mark the feed as read", func(t *testing.T) { markAsRead(t, intruder, "feed", feedID, now) }},
		{"mark the group as read", func(t *testing.T) { markAsRead(t, intruder, "group", category.ID, now) }},
		{"mark all items as read before now", func(t *testing.T) { markAsRead(t, intruder, "group", 0, now) }},
		{"mark all items as read", func(t *testing.T) { markAsRead(t, intruder, "group", 0, 0) }},
	}

	// The scenarios share the data of the owner, so the first change stops the test.
	for _, scenario := range scenarios {
		passed := t.Run(scenario.name, func(t *testing.T) {
			scenario.send(t)

			requireUnreadEntries(t, owner, entryIDs(entries)...)
			requireEntry(t, owner, entries[0].ID, miniflux.EntryStatusUnread, false)
			requireEntry(t, owner, entries[1].ID, miniflux.EntryStatusUnread, true)
		})
		if !passed {
			return
		}
	}
}

func formatIDs(ids []int64) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, formatID(id))
	}
	return values
}
