// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader_test

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestEditTagEndpoint(t *testing.T) {
	t.Parallel()

	read, unread := miniflux.EntryStatusRead, miniflux.EntryStatusUnread

	scenarios := []struct {
		name            string
		initialStatus   string
		initialStarred  bool
		form            url.Values
		expectedStatus  string
		expectedStarred bool
	}{
		{"add read", unread, false, url.Values{"a": {readStream}}, read, false},
		{"remove read", read, false, url.Values{"r": {readStream}}, unread, false},
		{"add kept-unread", read, false, url.Values{"a": {keptUnreadStream}}, unread, false},
		{"remove kept-unread", unread, false, url.Values{"r": {keptUnreadStream}}, read, false},
		{"add starred", unread, false, url.Values{"a": {starredStream}}, unread, true},
		{"remove starred", unread, true, url.Values{"r": {starredStream}}, unread, false},
		{"add read and starred", unread, false, url.Values{"a": {readStream, starredStream}}, read, true},
		{"add read and remove starred", unread, true, url.Values{"a": {readStream}, "r": {starredStream}}, read, false},
		{"like and broadcast are ignored", unread, false, url.Values{"a": {"user/-/state/com.google/like", "user/-/state/com.google/broadcast"}}, unread, false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			user := newTestUser(t)
			feedID, entries := subscribeToFixture(t, user)
			targetIDs := []int64{entries[0].ID, entries[1].ID}

			if scenario.initialStatus == read {
				if err := user.rest.UpdateEntries(targetIDs, read); err != nil {
					t.Fatal(err)
				}
			}
			if scenario.initialStarred {
				if err := user.rest.UpdateEntriesStarred(targetIDs, true); err != nil {
					t.Fatal(err)
				}
			}

			form := url.Values{"i": {strconv.FormatInt(entries[0].ID, 10), longItemID(entries[1].ID)}}
			for key, list := range scenario.form {
				form[key] = list
			}
			user.reader.postOK(t, "/reader/api/0/edit-tag", form)

			entries = feedEntries(t, user, feedID)
			for _, entry := range entries[:2] {
				if entry.Status != scenario.expectedStatus || entry.Starred != scenario.expectedStarred {
					t.Errorf(`Expected entry %d to be %s with starred=%v, got %s with starred=%v`, entry.ID, scenario.expectedStatus, scenario.expectedStarred, entry.Status, entry.Starred)
				}
			}
			if entries[2].Status != unread || entries[2].Starred {
				t.Errorf(`Expected the entry not listed in the request to be unchanged, got %s with starred=%v`, entries[2].Status, entries[2].Starred)
			}
		})
	}
}

func TestEditTagRejectsInvalidRequests(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	feedID, entries := subscribeToFixture(t, user)
	itemID := strconv.FormatInt(entries[0].ID, 10)

	scenarios := []struct {
		name           string
		form           url.Values
		expectedStatus int
	}{
		{"missing item ID", url.Values{"a": {readStream}}, http.StatusBadRequest},
		{"invalid item ID", url.Values{"i": {"invalid"}, "a": {readStream}}, http.StatusBadRequest},

		// The handler reports invalid tags as server errors.
		{"no tags", url.Values{"i": {itemID}}, http.StatusInternalServerError},
		{"invalid tag", url.Values{"i": {itemID}, "a": {"invalid"}}, http.StatusInternalServerError},
		{"label tag", url.Values{"i": {itemID}, "a": {labelStream("All")}}, http.StatusInternalServerError},
		{"add and remove read", url.Values{"i": {itemID}, "a": {readStream}, "r": {readStream}}, http.StatusInternalServerError},
		{"add read then kept-unread", url.Values{"i": {itemID}, "a": {readStream, keptUnreadStream}}, http.StatusInternalServerError},
		{"add kept-unread then read", url.Values{"i": {itemID}, "a": {keptUnreadStream, readStream}}, http.StatusInternalServerError},
		{"add and remove kept-unread", url.Values{"i": {itemID}, "a": {keptUnreadStream}, "r": {keptUnreadStream}}, http.StatusInternalServerError},
		{"add and remove starred", url.Values{"i": {itemID}, "a": {starredStream}, "r": {starredStream}}, http.StatusInternalServerError},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			status, body := user.reader.post(t, "/reader/api/0/edit-tag", scenario.form)
			requireStatus(t, status, body, scenario.expectedStatus)
		})
	}

	for _, entry := range feedEntries(t, user, feedID) {
		if entry.Status != miniflux.EntryStatusUnread || entry.Starred {
			t.Errorf(`Expected entry %d to be unchanged, got %s with starred=%v`, entry.ID, entry.Status, entry.Starred)
		}
	}
}

func TestMarkAllAsReadEndpoint(t *testing.T) {
	t.Parallel()

	read, unread := miniflux.EntryStatusRead, miniflux.EntryStatusUnread

	// The timestamp falls between the second and the third entries.
	before := time.Date(2025, time.January, 2, 12, 0, 0, 0, time.UTC)

	scenarios := []struct {
		name             string
		form             func(feedID int64) url.Values
		expectedStatuses []string
		expectedOthers   []string
	}{
		{
			"feed",
			func(feedID int64) url.Values { return url.Values{"s": {feedStream(feedID)}} },
			[]string{read, read, read},
			[]string{unread, unread, unread},
		},
		{
			"label",
			func(feedID int64) url.Values { return url.Values{"s": {labelStream("News")}} },
			[]string{read, read, read},
			[]string{unread, unread, unread},
		},
		{
			"reading list",
			func(feedID int64) url.Values { return url.Values{"s": {readingListStream}} },
			[]string{read, read, read},
			[]string{read, read, read},
		},
		{
			"feed before a timestamp in seconds",
			func(feedID int64) url.Values {
				return url.Values{"s": {feedStream(feedID)}, "ts": {strconv.FormatInt(before.Unix(), 10)}}
			},
			[]string{read, read, unread},
			[]string{unread, unread, unread},
		},
		{
			"feed before a timestamp in microseconds",
			func(feedID int64) url.Values {
				return url.Values{"s": {feedStream(feedID)}, "ts": {strconv.FormatInt(before.UnixMicro(), 10)}}
			},
			[]string{read, read, unread},
			[]string{unread, unread, unread},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			// The feed under test is in the News category, the other feed in All.
			user := newTestUser(t)
			category := createCategory(t, user, "News")
			feedID, err := user.rest.CreateFeed(&miniflux.FeedCreationRequest{
				FeedURL:    servertest.NewFeedFixture(t).FeedURL,
				CategoryID: category.ID,
			})
			if err != nil {
				t.Fatal(err)
			}
			otherFeedID, _ := subscribeToFixture(t, user)

			user.reader.postOK(t, "/reader/api/0/mark-all-as-read", scenario.form(feedID))

			if statuses := entryStatuses(t, user, feedID); !slices.Equal(statuses, scenario.expectedStatuses) {
				t.Errorf(`Unexpected statuses for the feed, got %v instead of %v`, statuses, scenario.expectedStatuses)
			}
			if statuses := entryStatuses(t, user, otherFeedID); !slices.Equal(statuses, scenario.expectedOthers) {
				t.Errorf(`Unexpected statuses for the other feed, got %v instead of %v`, statuses, scenario.expectedOthers)
			}
		})
	}
}

func TestMarkAllAsReadRejectsInvalidRequests(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	feedID, _ := subscribeToFixture(t, user)

	scenarios := []struct {
		name           string
		form           url.Values
		expectedStatus int
	}{
		{"unknown label", url.Values{"s": {labelStream("Unknown")}}, http.StatusNotFound},
		{"invalid stream", url.Values{"s": {"invalid"}}, http.StatusBadRequest},
		{"invalid feed ID", url.Values{"s": {"feed/invalid"}}, http.StatusBadRequest},
		{"invalid timestamp", url.Values{"s": {feedStream(feedID)}, "ts": {"yesterday"}}, http.StatusBadRequest},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			status, body := user.reader.post(t, "/reader/api/0/mark-all-as-read", scenario.form)
			requireStatus(t, status, body, scenario.expectedStatus)
		})
	}

	unread := []string{miniflux.EntryStatusUnread, miniflux.EntryStatusUnread, miniflux.EntryStatusUnread}
	if statuses := entryStatuses(t, user, feedID); !slices.Equal(statuses, unread) {
		t.Errorf(`Expected the entries to stay unread, got %v`, statuses)
	}
}

// entryStatuses returns the statuses of the feed entries oldest first, as seen by the REST API.
func entryStatuses(t *testing.T, user *testUser, feedID int64) []string {
	t.Helper()

	entries := feedEntries(t, user, feedID)
	statuses := make([]string, 0, len(entries))
	for _, entry := range entries {
		statuses = append(statuses, entry.Status)
	}
	return statuses
}
