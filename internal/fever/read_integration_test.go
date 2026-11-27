// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever_test

import (
	"cmp"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestGroups(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	defaultFeedID, _ := subscribeToFixture(t, user)
	defaultCategory := feedCategory(t, user, defaultFeedID)

	group := createCategory(t, user, "Fever Group")
	emptyGroup := createCategory(t, user, "Empty Fever Group")
	groupFeedIDs := sortedIDs(subscribeInCategory(t, user, group.ID), subscribeInCategory(t, user, group.ID))

	var result groupsResult
	user.fever.getJSON(t, selector("groups"), &result)

	expectedGroups := []groupResult{
		{ID: defaultCategory.ID, Title: defaultCategory.Title},
		{ID: group.ID, Title: "Fever Group"},
		{ID: emptyGroup.ID, Title: "Empty Fever Group"},
	}
	requireGroups(t, result.Groups, expectedGroups)

	// A group without feeds is not in feeds_groups.
	requireFeedsGroups(t, result.FeedsGroups, map[int64][]int64{
		defaultCategory.ID: {defaultFeedID},
		group.ID:           groupFeedIDs,
	})
}

func TestGroupsWithoutFeeds(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	var result groupsResult
	user.fever.getJSON(t, selector("groups"), &result)

	if len(result.Groups) != 1 {
		t.Errorf(`Expected the default category only, got %+v`, result.Groups)
	}
	if result.FeedsGroups == nil || len(result.FeedsGroups) != 0 {
		t.Errorf(`Expected an empty feeds_groups array, got %+v`, result.FeedsGroups)
	}
}

func TestFeeds(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	fixture := servertest.NewFeedFixture(t)
	feedID := fixture.Subscribe(t, user.rest)

	feed, err := user.rest.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}
	icon, err := user.rest.FeedIcon(feedID)
	if err != nil {
		t.Fatal(err)
	}

	var result feedsResult
	user.fever.getJSON(t, selector("feeds"), &result)

	// Miniflux has no sparks: every feed is in the "Kindling" super group.
	expected := []feedResult{{
		ID:                feedID,
		FaviconID:         icon.ID,
		Title:             servertest.FeedTitle,
		URL:               fixture.FeedURL,
		SiteURL:           fixture.WebsiteURL,
		IsSpark:           0,
		LastUpdatedOnTime: feed.CheckedAt.Unix(),
	}}
	if !slices.Equal(result.Feeds, expected) {
		t.Errorf(`Unexpected feeds, got %+v instead of %+v`, result.Feeds, expected)
	}

	requireFeedsGroups(t, result.FeedsGroups, map[int64][]int64{feed.Category.ID: {feedID}})
}

func TestFeedsWithoutFeeds(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	var result feedsResult
	user.fever.getJSON(t, selector("feeds"), &result)

	if result.Feeds == nil || len(result.Feeds) != 0 {
		t.Errorf(`Expected an empty feeds array, got %+v`, result.Feeds)
	}
	if result.FeedsGroups == nil || len(result.FeedsGroups) != 0 {
		t.Errorf(`Expected an empty feeds_groups array, got %+v`, result.FeedsGroups)
	}
}

func TestFavicons(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	feedID, _ := subscribeToFixture(t, user)

	icon, err := user.rest.FeedIcon(feedID)
	if err != nil {
		t.Fatal(err)
	}

	var result faviconsResult
	user.fever.getJSON(t, selector("favicons"), &result)

	expected := []faviconResult{{ID: icon.ID, Data: icon.Data}}
	if !slices.Equal(result.Favicons, expected) {
		t.Fatalf(`Unexpected favicons, got %+v instead of %+v`, result.Favicons, expected)
	}

	// Clients use the data with the data: scheme, so it starts with the MIME type.
	if !strings.HasPrefix(result.Favicons[0].Data, "image/png;base64,") {
		t.Errorf(`Expected PNG data with a MIME type prefix, got %q`, result.Favicons[0].Data)
	}
}

func TestUnreadAndSavedItemIDs(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	_, entries := subscribeToFixture(t, user)

	requireUnreadItemIDs(t, user, entries[0].ID, entries[1].ID, entries[2].ID)
	requireSavedItemIDs(t, user)

	// The saved items are the starred entries, read or not.
	if err := user.rest.UpdateEntries([]int64{entries[0].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries[:2] {
		if err := user.rest.ToggleStarred(entry.ID); err != nil {
			t.Fatal(err)
		}
	}

	requireUnreadItemIDs(t, user, entries[1].ID, entries[2].ID)
	requireSavedItemIDs(t, user, entries[0].ID, entries[1].ID)
}

func TestReadSelectorsInPOSTRequests(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	subscribeToFixture(t, user)

	// Fever clients send the selector in the URL and the API key as POST data.
	for _, name := range []string{"groups", "feeds", "favicons", "unread_item_ids", "saved_item_ids", "items"} {
		t.Run(name, func(t *testing.T) {
			status, body := user.fever.postWithQuery(t, selector(name), nil)

			var result map[string]json.RawMessage
			decodeAuthenticatedJSON(t, status, body, &result)
			if _, found := result[name]; !found {
				t.Errorf(`Expected the %s member, got %s`, name, body)
			}
		})
	}
}

// feedCategory returns the category of the feed, as seen by the REST API.
func feedCategory(t *testing.T, user *testUser, feedID int64) *miniflux.Category {
	t.Helper()

	feed, err := user.rest.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}
	return feed.Category
}

// requireGroups fails the test unless the groups are the expected ones, in any order.
func requireGroups(t *testing.T, groups, expected []groupResult) {
	t.Helper()

	byID := func(a, b groupResult) int { return cmp.Compare(a.ID, b.ID) }
	groups = slices.SortedFunc(slices.Values(groups), byID)
	expected = slices.SortedFunc(slices.Values(expected), byID)
	if !slices.Equal(groups, expected) {
		t.Errorf(`Unexpected groups, got %+v instead of %+v`, groups, expected)
	}
}

// requireFeedsGroups fails the test unless feeds_groups maps each group ID to the expected feed IDs, in any order.
func requireFeedsGroups(t *testing.T, feedsGroups []feedsGroupResult, expected map[int64][]int64) {
	t.Helper()

	feedIDsByGroup := make(map[int64][]int64, len(feedsGroups))
	for _, feedsGroup := range feedsGroups {
		if _, found := feedIDsByGroup[feedsGroup.GroupID]; found {
			t.Errorf(`Group %d is in feeds_groups more than once: %+v`, feedsGroup.GroupID, feedsGroups)
		}
		feedIDsByGroup[feedsGroup.GroupID] = parseIDs(t, feedsGroup.FeedIDs)
	}
	if !reflect.DeepEqual(feedIDsByGroup, expected) {
		t.Errorf(`Unexpected feeds_groups, got %v instead of %v`, feedIDsByGroup, expected)
	}
}

// requireUnreadItemIDs fails the test unless unread_item_ids lists the expected entries, in any order.
func requireUnreadItemIDs(t *testing.T, user *testUser, expected ...int64) {
	t.Helper()

	var result unreadItemIDsResult
	user.fever.getJSON(t, selector("unread_item_ids"), &result)
	if ids := parseIDs(t, result.UnreadItemIDs); !slices.Equal(ids, sortedIDs(expected...)) {
		t.Errorf(`Expected unread items %v, got %v`, sortedIDs(expected...), ids)
	}
}

// requireSavedItemIDs fails the test unless saved_item_ids lists the expected entries, in any order.
func requireSavedItemIDs(t *testing.T, user *testUser, expected ...int64) {
	t.Helper()

	var result savedItemIDsResult
	user.fever.getJSON(t, selector("saved_item_ids"), &result)
	if ids := parseIDs(t, result.SavedItemIDs); !slices.Equal(ids, sortedIDs(expected...)) {
		t.Errorf(`Expected saved items %v, got %v`, sortedIDs(expected...), ids)
	}
}
