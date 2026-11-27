// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever_test

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
)

// beforeNewestEntry is a before timestamp between the publication of fixture entries 2 and 3.
var beforeNewestEntry = time.Date(2025, time.January, 2, 12, 0, 0, 0, time.UTC).Unix()

func TestMarkItem(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	_, entries := subscribeToFixture(t, user)
	entryID := entries[0].ID

	// The steps share the entry, so each one starts from the state left by the previous one.
	steps := []struct {
		as      string
		status  string
		starred bool
	}{
		{"read", miniflux.EntryStatusRead, false},
		{"unread", miniflux.EntryStatusUnread, false},
		{"saved", miniflux.EntryStatusUnread, true},
		{"saved", miniflux.EntryStatusUnread, true},
		{"unsaved", miniflux.EntryStatusUnread, false},
		{"unsaved", miniflux.EntryStatusUnread, false},
		{"unknown", miniflux.EntryStatusUnread, false},
		{"", miniflux.EntryStatusUnread, false},
	}

	for index, step := range steps {
		passed := t.Run(fmt.Sprintf("step %d as=%q", index+1, step.as), func(t *testing.T) {
			markItem(t, user.fever.postJSON, entryID, step.as)
			requireEntry(t, user, entryID, step.status, step.starred)
		})
		if !passed {
			return
		}
	}

	// Only the marked entry changed.
	for _, entry := range entries[1:] {
		requireEntry(t, user, entry.ID, miniflux.EntryStatusUnread, false)
	}
}

func TestMarkItemWithParametersInQueryString(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	_, entries := subscribeToFixture(t, user)

	markItem(t, user.fever.getJSON, entries[0].ID, "read")
	requireEntry(t, user, entries[0].ID, miniflux.EntryStatusRead, false)

	markItem(t, user.fever.getJSON, entries[0].ID, "saved")
	requireEntry(t, user, entries[0].ID, miniflux.EntryStatusRead, true)
}

func TestMarkFeed(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	feedID, entries := subscribeToFixture(t, user)
	_, otherEntries := subscribeToFixture(t, user)
	allEntries := append(slices.Clone(entries), otherEntries...)

	// Without before, the timestamp is 0 and no entry is older.
	markAsRead(t, user, "feed", feedID, 0)
	requireUnreadEntries(t, user, entryIDs(allEntries)...)

	markAsRead(t, user, "feed", feedID, beforeNewestEntry)
	requireUnreadEntries(t, user, entryIDs(append([]*miniflux.Entry{entries[2]}, otherEntries...))...)
}

func TestMarkGroup(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	group := createCategory(t, user, "Fever Group")
	groupEntries := feedEntries(t, user, subscribeInCategory(t, user, group.ID))
	_, otherEntries := subscribeToFixture(t, user)
	allEntries := append(slices.Clone(groupEntries), otherEntries...)

	// Without before, the timestamp is 0 and no entry is older.
	markAsRead(t, user, "group", group.ID, 0)
	requireUnreadEntries(t, user, entryIDs(allEntries)...)

	markAsRead(t, user, "group", group.ID, beforeNewestEntry)
	requireUnreadEntries(t, user, entryIDs(append([]*miniflux.Entry{groupEntries[2]}, otherEntries...))...)
}

func TestMarkAllItems(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	group := createCategory(t, user, "Fever Group")
	groupEntries := feedEntries(t, user, subscribeInCategory(t, user, group.ID))
	_, otherEntries := subscribeToFixture(t, user)

	// Group 0 is the "Kindling" super group, which contains every feed.
	markAsRead(t, user, "group", 0, beforeNewestEntry)
	requireUnreadEntries(t, user, groupEntries[2].ID, otherEntries[2].ID)

	// Without before, every entry is marked as read.
	markAsRead(t, user, "group", 0, 0)
	requireUnreadEntries(t, user)
}

// markItem sends a mark=item request with the given sender and requires the authenticated response.
// An empty as is left out of the request.
func markItem(t *testing.T, send func(*testing.T, url.Values, any), entryID int64, as string) {
	t.Helper()

	params := url.Values{"mark": {"item"}, "id": {formatID(entryID)}}
	if as != "" {
		params.Set("as", as)
	}

	var result baseResult
	send(t, params, &result)
}

// markAsRead sends a mark=feed or mark=group request as POST data and requires the authenticated response.
// A zero before is left out of the request.
func markAsRead(t *testing.T, user *testUser, mark string, id, before int64) {
	t.Helper()

	params := url.Values{"mark": {mark}, "as": {"read"}, "id": {formatID(id)}}
	if before != 0 {
		params.Set("before", strconv.FormatInt(before, 10))
	}

	var result baseResult
	user.fever.postJSON(t, params, &result)
}

// requireEntry fails the test unless the REST API reports the expected status and starred flag for the entry.
func requireEntry(t *testing.T, user *testUser, entryID int64, status string, starred bool) {
	t.Helper()

	entry, err := user.rest.Entry(entryID)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Status != status || entry.Starred != starred {
		t.Fatalf(`Expected entry %d to be %s with starred=%v, got %s with starred=%v`, entryID, status, starred, entry.Status, entry.Starred)
	}
}

// requireUnreadEntries fails the test unless the REST API reports exactly the expected unread entries.
func requireUnreadEntries(t *testing.T, user *testUser, expected ...int64) {
	t.Helper()

	result, err := user.rest.Entries(&miniflux.Filter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		t.Fatal(err)
	}
	if ids := sortedIDs(entryIDs(result.Entries)...); !slices.Equal(ids, sortedIDs(expected...)) {
		t.Fatalf(`Expected unread entries %v, got %v`, sortedIDs(expected...), ids)
	}
}

func entryIDs(entries miniflux.Entries) []int64 {
	ids := make([]int64, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}
	return ids
}
