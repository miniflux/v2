// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader_test

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestStreamItemIDsEndpoint(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	feedID, entries := subscribeToFixture(t, user)
	_, otherEntries := subscribeToFixture(t, user)

	if err := user.rest.UpdateEntries([]int64{entries[0].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}
	if err := user.rest.UpdateEntriesStarred([]int64{entries[2].ID}, true); err != nil {
		t.Fatal(err)
	}

	allEntries := append(slices.Clone(entries), otherEntries...)
	secondDay := time.Date(2025, time.January, 2, 0, 0, 0, 0, time.UTC)

	scenarios := []struct {
		name  string
		query url.Values
		// expected lists the entries in the order of the response, unless unordered is set:
		// the entries of both feeds are published at the same times.
		expected  miniflux.Entries
		unordered bool
	}{
		{"feed newest first", url.Values{"s": {feedStream(feedID)}}, miniflux.Entries{entries[2], entries[1], entries[0]}, false},
		{"feed oldest first", url.Values{"s": {feedStream(feedID)}, "r": {"o"}}, entries, false},
		{"feed without read items", url.Values{"s": {feedStream(feedID)}, "xt": {readStream}}, miniflux.Entries{entries[2], entries[1]}, false},
		{"reading list", url.Values{"s": {readingListStream}}, allEntries, true},
		{"reading list without read items", url.Values{"s": {readingListStream}, "xt": {readStream}}, allEntries[1:], true},
		{"reading list published during the second day", url.Values{
			"s":  {readingListStream},
			"ot": {strconv.FormatInt(secondDay.Unix(), 10)},
			"nt": {strconv.FormatInt(secondDay.Add(24*time.Hour).Unix(), 10)},
		}, miniflux.Entries{entries[1], otherEntries[1]}, true},
		{"starred", url.Values{"s": {starredStream}}, miniflux.Entries{entries[2]}, false},
		{"starred with the user ID", url.Values{"s": {fmt.Sprintf("user/%d/state/com.google/starred", user.id)}}, miniflux.Entries{entries[2]}, false},
		{"read", url.Values{"s": {readStream}}, miniflux.Entries{entries[0]}, false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			var result itemIDsResult
			user.reader.getJSON(t, "/reader/api/0/stream/items/ids", scenario.query, &result)

			ids := result.ids()
			expected := decimalItemIDs(scenario.expected)
			if scenario.unordered {
				slices.Sort(ids)
				slices.Sort(expected)
			}
			if !slices.Equal(ids, expected) {
				t.Errorf(`Unexpected item IDs, got %v instead of %v`, ids, expected)
			}
			if result.Continuation != "" {
				t.Errorf(`Expected no continuation, got %q`, result.Continuation)
			}
		})
	}
}

func TestStreamItemIDsPagination(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	feedID, entries := subscribeToFixture(t, user)

	var firstPage itemIDsResult
	user.reader.getJSON(t, "/reader/api/0/stream/items/ids", url.Values{"s": {feedStream(feedID)}, "n": {"2"}}, &firstPage)

	if expected := decimalItemIDs(miniflux.Entries{entries[2], entries[1]}); !slices.Equal(firstPage.ids(), expected) {
		t.Errorf(`Unexpected first page, got %v instead of %v`, firstPage.ids(), expected)
	}
	if firstPage.Continuation != "2" {
		t.Fatalf(`Expected the continuation "2", got %q`, firstPage.Continuation)
	}

	var lastPage itemIDsResult
	user.reader.getJSON(t, "/reader/api/0/stream/items/ids", url.Values{
		"s": {feedStream(feedID)},
		"n": {"2"},
		"c": {firstPage.Continuation},
	}, &lastPage)

	if expected := decimalItemIDs(miniflux.Entries{entries[0]}); !slices.Equal(lastPage.ids(), expected) {
		t.Errorf(`Unexpected last page, got %v instead of %v`, lastPage.ids(), expected)
	}
	if lastPage.Continuation != "" {
		t.Errorf(`Expected no continuation on the last page, got %q`, lastPage.Continuation)
	}
}

func TestStreamItemIDsRejectsInvalidStreams(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	scenarios := []struct {
		name  string
		query url.Values
	}{
		{"missing stream", url.Values{}},
		{"invalid stream", url.Values{"s": {"invalid"}}},
		{"invalid feed ID", url.Values{"s": {"feed/invalid"}}},
		{"zero feed ID", url.Values{"s": {"feed/0"}}},
		{"negative feed ID", url.Values{"s": {"feed/-1"}}},
		{"invalid exclusion", url.Values{"s": {readingListStream}, "xt": {"invalid"}}},
		{"two streams", url.Values{"s": {readingListStream, starredStream}}},

		// Label streams are not supported.
		{"label stream", url.Values{"s": {labelStream("All")}}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			query := url.Values{"output": {"json"}}
			for key, list := range scenario.query {
				query[key] = list
			}

			status, body := user.reader.get(t, "/reader/api/0/stream/items/ids", query)
			requireStatus(t, status, body, http.StatusBadRequest)
		})
	}
}

func TestStreamItemContentsEndpoint(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	fixture := servertest.NewFeedFixture(t)
	category := createCategory(t, user, "News")

	feedID, err := user.rest.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: fixture.FeedURL, CategoryID: category.ID})
	if err != nil {
		t.Fatal(err)
	}
	entries := feedEntries(t, user, feedID)

	if err := user.rest.UpdateEntries([]int64{entries[0].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}
	if err := user.rest.UpdateEntriesStarred([]int64{entries[1].ID}, true); err != nil {
		t.Fatal(err)
	}

	// Clients send item IDs in decimal, in the long form, or as 16 hexadecimal digits.
	var result itemContentsResult
	user.reader.postJSON(t, "/reader/api/0/stream/items/contents", url.Values{"i": {
		strconv.FormatInt(entries[0].ID, 10),
		longItemID(entries[1].ID),
		fmt.Sprintf("%016x", entries[2].ID),
	}}, &result)

	if result.ID != readingListStream || result.Author != user.username {
		t.Errorf(`Unexpected stream %q by %q`, result.ID, result.Author)
	}

	userPrefix := fmt.Sprintf("user/%d/", user.id)
	readingList := userPrefix + "state/com.google/reading-list"
	label := userPrefix + "label/News"
	expectedCategories := [][]string{
		{readingList, label, userPrefix + "state/com.google/read"},
		{readingList, label, userPrefix + "state/com.google/starred"},
		{readingList, label},
	}
	expectedEnclosures := [][]itemEnclosureResult{
		{},
		{{URL: fixture.WebsiteURL + "media/episode-2.mp3", Type: "audio/mpeg"}},
		{},
	}

	// The items are sorted newest first.
	expected := make([]itemResult, 0, len(entries))
	for index := len(entries) - 1; index >= 0; index-- {
		entry := entries[index]
		expected = append(expected, itemResult{
			ID:         longItemID(entry.ID),
			Title:      entry.Title,
			Categories: expectedCategories[index],
			Published:  entry.Date.Unix(),
			Canonical:  []hrefResult{{HREF: entry.URL}},
			Alternate:  []hrefResult{{HREF: entry.URL, Type: "text/html"}},
			Content:    itemContentResult{Content: entry.Content},
			Origin:     itemOriginResult{StreamID: feedStream(feedID), Title: servertest.FeedTitle, HTMLURL: fixture.WebsiteURL},
			Enclosure:  expectedEnclosures[index],
		})
	}

	if len(result.Items) != len(expected) {
		t.Fatalf(`Expected %d items, got %d`, len(expected), len(result.Items))
	}
	for index := range expected {
		if !reflect.DeepEqual(result.Items[index], expected[index]) {
			t.Errorf(`Unexpected item %d, got %+v instead of %+v`, index, result.Items[index], expected[index])
		}
	}
}

func TestStreamItemContentsRejectsInvalidRequests(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	// The stream filters are read from the query string, the item IDs from the form.
	scenarios := []struct {
		name    string
		query   string
		itemIDs []string
	}{
		{"missing item ID", "", nil},
		{"invalid item ID", "", []string{"invalid"}},
		{"zero long form item ID", "", []string{longItemID(0)}},
		{"invalid exclusion", "?xt=invalid", []string{"1"}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			status, body := user.reader.post(t, "/reader/api/0/stream/items/contents"+scenario.query, url.Values{
				"output": {"json"},
				"i":      scenario.itemIDs,
			})
			requireStatus(t, status, body, http.StatusBadRequest)
		})
	}
}

func TestEndpointsRejectOutputOtherThanJSON(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	for _, output := range []string{"", "xml"} {
		for _, path := range []string{"/reader/api/0/tag/list", "/reader/api/0/subscription/list", "/reader/api/0/stream/items/ids"} {
			t.Run("GET "+path+" with output "+strconv.Quote(output), func(t *testing.T) {
				status, body := user.reader.get(t, path, url.Values{"output": {output}, "s": {readingListStream}})
				requireStatus(t, status, body, http.StatusBadRequest)
			})
		}

		t.Run("POST /reader/api/0/stream/items/contents with output "+strconv.Quote(output), func(t *testing.T) {
			status, body := user.reader.post(t, "/reader/api/0/stream/items/contents", url.Values{"output": {output}, "i": {"1"}})
			requireStatus(t, status, body, http.StatusBadRequest)
		})
	}
}

// decimalItemIDs returns the item IDs of the entries in the decimal form used by stream/items/ids.
func decimalItemIDs(entries miniflux.Entries) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, strconv.FormatInt(entry.ID, 10))
	}
	return ids
}
