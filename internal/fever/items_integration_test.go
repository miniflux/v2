// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever_test

import (
	"cmp"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestItems(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	feedID, entries := subscribeToFixture(t, user)

	if err := user.rest.UpdateEntries([]int64{entries[0].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}
	if err := user.rest.ToggleStarred(entries[1].ID); err != nil {
		t.Fatal(err)
	}

	var result itemsResult
	user.fever.getJSON(t, selector("items"), &result)

	if result.TotalItems != servertest.EntryCount {
		t.Errorf(`Expected total_items %d, got %d`, servertest.EntryCount, result.TotalItems)
	}

	// Items are sorted by ID. Entry 1 is read and entry 2 is saved.
	expected := make([]itemResult, 0, len(entries))
	for index, entry := range entries {
		expected = append(expected, itemResult{
			ID:            entry.ID,
			FeedID:        feedID,
			Title:         entry.Title,
			Author:        entry.Author,
			HTML:          entry.Content,
			URL:           entry.URL,
			IsSaved:       boolInt(index == 1),
			IsRead:        boolInt(index == 0),
			CreatedOnTime: entry.Date.Unix(),
		})
	}
	slices.SortFunc(expected, func(left, right itemResult) int { return cmp.Compare(left.ID, right.ID) })

	if !slices.Equal(result.Items, expected) {
		t.Fatalf(`Unexpected items, got %+v instead of %+v`, result.Items, expected)
	}

	// The fields come from the fixture feed, not only from the REST API.
	item := result.Items[slices.IndexFunc(result.Items, func(item itemResult) bool { return item.ID == entries[1].ID })]
	if item.Title != "Entry 2" || !strings.Contains(item.HTML, servertest.SearchTerm) {
		t.Errorf(`Expected the title and content of entry 2, got %+v`, item)
	}
	if expectedTime := time.Date(2025, time.January, 2, 10, 0, 0, 0, time.UTC).Unix(); item.CreatedOnTime != expectedTime {
		t.Errorf(`Expected created_on_time %d, got %d`, expectedTime, item.CreatedOnTime)
	}
}

func TestItemsSelection(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	_, entries := subscribeToFixture(t, user)

	ids := sortedIDs(entryIDs(entries)...)
	oldest, middle, newest := ids[0], ids[1], ids[2]

	scenarios := []struct {
		name     string
		params   url.Values
		expected []int64
	}{
		{"no argument returns the oldest items first", nil, []int64{oldest, middle, newest}},
		{"since_id returns the newer items, oldest first", url.Values{"since_id": {formatID(oldest)}}, []int64{middle, newest}},
		{"since_id=0 returns the oldest items first", url.Values{"since_id": {"0"}}, []int64{oldest, middle, newest}},
		{"since_id of the newest item returns nothing", url.Values{"since_id": {formatID(newest)}}, []int64{}},
		{"max_id returns the older items, newest first", url.Values{"max_id": {formatID(newest)}}, []int64{middle, oldest}},
		{"max_id=0 returns the newest items first", url.Values{"max_id": {"0"}}, []int64{newest, middle, oldest}},
		{"max_id of the oldest item returns nothing", url.Values{"max_id": {formatID(oldest)}}, []int64{}},
		{"with_ids returns the items, oldest first", url.Values{"with_ids": {formatID(newest) + "," + formatID(oldest)}}, []int64{oldest, newest}},
		{"with_ids ignores spaces", url.Values{"with_ids": {formatID(newest) + ", " + formatID(oldest)}}, []int64{oldest, newest}},
		{"with_ids ignores invalid IDs", url.Values{"with_ids": {formatID(middle) + ",invalid"}}, []int64{middle}},
		{"with_ids without IDs returns nothing", url.Values{"with_ids": {""}}, []int64{}},
		{"with_ids without valid IDs returns nothing", url.Values{"with_ids": {"invalid,0,-1"}}, []int64{}},
		{
			"since_id takes precedence over max_id",
			url.Values{"since_id": {formatID(oldest)}, "max_id": {formatID(middle)}},
			[]int64{middle, newest},
		},
		{
			"max_id takes precedence over with_ids",
			url.Values{"max_id": {formatID(newest)}, "with_ids": {formatID(newest)}},
			[]int64{middle, oldest},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			query := selector("items")
			for key, values := range scenario.params {
				query[key] = values
			}

			var result itemsResult
			user.fever.getJSON(t, query, &result)

			if result.Items == nil {
				t.Fatalf(`Expected an items array, got null`)
			}
			if !slices.Equal(result.ids(), scenario.expected) {
				t.Errorf(`Expected items %v, got %v`, scenario.expected, result.ids())
			}

			// total_items counts every item of the user, whatever the selection.
			if result.TotalItems != servertest.EntryCount {
				t.Errorf(`Expected total_items %d, got %d`, servertest.EntryCount, result.TotalItems)
			}
		})
	}
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

// boolInt returns the Fever boolean integer: 1 for true, 0 for false.
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
