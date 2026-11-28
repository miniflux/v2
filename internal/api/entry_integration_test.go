// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
	"miniflux.app/v2/internal/model"
)

func TestGetAllFeedEntriesEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	results := listFeedEntries(t, regularUserClient, feedID, nil)

	if len(results.Entries) == 0 {
		t.Fatalf(`Invalid number of entries, got %d`, len(results.Entries))
	}

	if results.Total == 0 {
		t.Fatalf(`Invalid total, got %d`, results.Total)
	}

	if results.Entries[0].FeedID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, results.Entries[0].FeedID)
	}

	if results.Entries[0].Feed.FeedURL != fixture.FeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, results.Entries[0].Feed.FeedURL)
	}

	if results.Entries[0].Title == "" {
		t.Fatalf(`Invalid title, got empty`)
	}
}

func TestGetAllCategoryEntriesEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    fixture.FeedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := regularUserClient.CategoryEntries(category.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Entries) == 0 {
		t.Fatalf(`Invalid number of entries, got %d`, len(results.Entries))
	}

	if results.Total == 0 {
		t.Fatalf(`Invalid total, got %d`, results.Total)
	}

	if results.Entries[0].FeedID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, results.Entries[0].FeedID)
	}

	if results.Entries[0].Feed.FeedURL != fixture.FeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, results.Entries[0].Feed.FeedURL)
	}

	if results.Entries[0].Title == "" {
		t.Fatalf(`Invalid title, got empty`)
	}
}

func TestGetAllEntriesEndpointWithFilter(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	feedEntries, err := regularUserClient.Entries(&miniflux.Filter{FeedID: feedID})
	if err != nil {
		t.Fatal(err)
	}

	if len(feedEntries.Entries) != servertest.EntryCount {
		t.Fatalf(`Invalid number of entries, got %d instead of %d`, len(feedEntries.Entries), servertest.EntryCount)
	}

	if feedEntries.Total != servertest.EntryCount {
		t.Fatalf(`Invalid total, got %d instead of %d`, feedEntries.Total, servertest.EntryCount)
	}

	if feedEntries.Entries[0].FeedID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, feedEntries.Entries[0].FeedID)
	}

	if feedEntries.Entries[0].Feed.FeedURL != fixture.FeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, feedEntries.Entries[0].Feed.FeedURL)
	}

	// Entries are sorted by ascending publication date by default.
	if feedEntries.Entries[0].Title != "Entry 1" {
		t.Fatalf(`Invalid title of the oldest entry, got %q`, feedEntries.Entries[0].Title)
	}

	emptyPage, err := regularUserClient.Entries(&miniflux.Filter{
		FeedID: feedID,
		Limit:  1,
		Offset: feedEntries.Total,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(emptyPage.Entries) != 0 {
		t.Fatalf(`Expected no entries beyond the final page, got %d`, len(emptyPage.Entries))
	}

	if emptyPage.Total != feedEntries.Total {
		t.Fatalf(`Expected total %d beyond the final page, got %d`, feedEntries.Total, emptyPage.Total)
	}

	recentEntries, err := regularUserClient.Entries(&miniflux.Filter{Order: "published_at", Direction: "desc"})
	if err != nil {
		t.Fatal(err)
	}

	if len(recentEntries.Entries) != servertest.EntryCount {
		t.Fatalf(`Invalid number of entries, got %d instead of %d`, len(recentEntries.Entries), servertest.EntryCount)
	}

	if recentEntries.Entries[0].Title != "Entry 3" {
		t.Fatalf(`Invalid order, got %q instead of the most recent entry`, recentEntries.Entries[0].Title)
	}

	searchedEntries, err := regularUserClient.Entries(&miniflux.Filter{Search: servertest.SearchTerm})
	if err != nil {
		t.Fatal(err)
	}

	if searchedEntries.Total != 1 || len(searchedEntries.Entries) != 1 {
		t.Fatalf(`Invalid search results, got %d entries and total %d`, len(searchedEntries.Entries), searchedEntries.Total)
	}

	if searchedEntries.Entries[0].Title != "Entry 2" {
		t.Fatalf(`Invalid search result, got %q`, searchedEntries.Entries[0].Title)
	}

	if _, err := regularUserClient.Entries(&miniflux.Filter{Status: "invalid"}); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Using invalid status should raise an error, got %v`, err)
	}

	if _, err = regularUserClient.Entries(&miniflux.Filter{Direction: "invalid"}); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Using invalid direction should raise an error, got %v`, err)
	}

	if _, err = regularUserClient.Entries(&miniflux.Filter{Order: "invalid"}); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Using invalid order should raise an error, got %v`, err)
	}
}

func TestGetGlobalEntriesEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:      fixture.FeedURL,
		HideGlobally: true,
	})

	if err != nil {
		t.Fatal(err)
	}

	feedIDEntry, err := regularUserClient.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if feedIDEntry.HideGlobally != true {
		t.Fatalf(`Expected feed to have globally_hidden set to true, was false.`)
	}

	/* Not filtering on GloballyVisible should return all entries */
	feedEntries, err := regularUserClient.Entries(&miniflux.Filter{FeedID: feedID})

	if err != nil {
		t.Fatal(err)
	}

	if len(feedEntries.Entries) == 0 {
		t.Fatalf(`Expected entries but response contained none.`)
	}

	/* Feed is hidden globally, so this should be empty */
	globallyVisibleEntries, err := regularUserClient.Entries(&miniflux.Filter{GloballyVisible: true})

	if err != nil {
		t.Fatal(err)
	}

	if len(globallyVisibleEntries.Entries) != 0 {
		t.Fatalf(`Expected no entries, got %d`, len(globallyVisibleEntries.Entries))
	}
}

func TestGetEntryEndpoints(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, nil)

	entry, err := regularUserClient.FeedEntry(feedID, result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.ID != result.Entries[0].ID {
		t.Fatalf(`Invalid entryID, got %d`, entry.ID)
	}

	if entry.FeedID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, entry.FeedID)
	}

	if entry.Feed.FeedURL != fixture.FeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, entry.Feed.FeedURL)
	}

	entry, err = regularUserClient.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.ID != result.Entries[0].ID {
		t.Fatalf(`Invalid entryID, got %d`, entry.ID)
	}

	entry, err = regularUserClient.CategoryEntry(result.Entries[0].Feed.Category.ID, result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.ID != result.Entries[0].ID {
		t.Fatalf(`Invalid entryID, got %d`, entry.ID)
	}
}

func TestUpdateEntryStatusEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, nil)

	if err := regularUserClient.UpdateEntries([]int64{result.Entries[0].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}

	entry, err := regularUserClient.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Status != miniflux.EntryStatusRead {
		t.Fatalf(`Invalid status, got %q`, entry.Status)
	}
}

func TestUpdateEntriesStarredEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, nil)

	entryID := result.Entries[0].ID

	// Star the entry without changing its status.
	if err := regularUserClient.UpdateEntriesStarred([]int64{entryID}, true); err != nil {
		t.Fatal(err)
	}

	entry, err := regularUserClient.Entry(entryID)
	if err != nil {
		t.Fatal(err)
	}

	if !entry.Starred {
		t.Fatalf(`Expected entry to be starred`)
	}

	if entry.Status != miniflux.EntryStatusUnread {
		t.Fatalf(`Expected status to remain unread, got %q`, entry.Status)
	}

	// Unstar the entry.
	if err := regularUserClient.UpdateEntriesStarred([]int64{entryID}, false); err != nil {
		t.Fatal(err)
	}

	entry, err = regularUserClient.Entry(entryID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Starred {
		t.Fatalf(`Expected entry to no longer be starred`)
	}
}

func TestGetEntryIDsEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	boolPtr := func(value bool) *bool { return &value }

	// A new user should have no entries at all.
	result, err := regularUserClient.EntryIDs(nil)
	if err != nil {
		t.Fatal(err)
	}

	if result.EntryIDs == nil {
		t.Fatal(`Entry IDs should not be nil`)
	}

	if len(result.EntryIDs) != 0 {
		t.Fatalf(`Expected no entry IDs for a new user, got %d`, len(result.EntryIDs))
	}

	if result.Total != 0 {
		t.Fatalf(`Expected total to be 0 for a new user, got %d`, result.Total)
	}

	// Subscribe to a feed so there are entries.
	feedID := fixture.Subscribe(t, regularUserClient)

	allEntries := listFeedEntries(t, regularUserClient, feedID, nil)

	if len(allEntries.Entries) != servertest.EntryCount {
		t.Fatalf(`Expected %d entries in the feed, got %d`, servertest.EntryCount, len(allEntries.Entries))
	}

	// Without filters, all entries should be returned.
	result, err = regularUserClient.EntryIDs(nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.EntryIDs) != allEntries.Total {
		t.Fatalf(`Expected %d entry IDs, got %d`, allEntries.Total, len(result.EntryIDs))
	}

	if result.Total != allEntries.Total {
		t.Fatalf(`Expected total %d, got %d`, allEntries.Total, result.Total)
	}

	// Filter by status=unread: all entries should be unread initially.
	unreadResult, err := regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		t.Fatal(err)
	}

	if len(unreadResult.EntryIDs) != allEntries.Total {
		t.Fatalf(`Expected %d unread entry IDs, got %d`, allEntries.Total, len(unreadResult.EntryIDs))
	}

	// Mark one entry as read and verify status filter results update.
	firstEntryID := allEntries.Entries[0].ID
	if err := regularUserClient.UpdateEntries([]int64{firstEntryID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}

	unreadResult, err = regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		t.Fatal(err)
	}

	if len(unreadResult.EntryIDs) != allEntries.Total-1 {
		t.Fatalf(`Expected %d unread entry IDs after marking one as read, got %d`, allEntries.Total-1, len(unreadResult.EntryIDs))
	}

	if unreadResult.Total != allEntries.Total-1 {
		t.Fatalf(`Expected total %d after marking one as read, got %d`, allEntries.Total-1, unreadResult.Total)
	}

	for _, id := range unreadResult.EntryIDs {
		if id == firstEntryID {
			t.Fatalf(`Entry ID %d should not appear in unread IDs after being marked as read`, firstEntryID)
		}
	}

	readResult, err := regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Status: miniflux.EntryStatusRead})
	if err != nil {
		t.Fatal(err)
	}

	if len(readResult.EntryIDs) != 1 || readResult.EntryIDs[0] != firstEntryID {
		t.Fatalf(`Expected only entry %d in read results, got %v`, firstEntryID, readResult.EntryIDs)
	}

	// Pagination: limit=1 should return 1 entry but total reflects the full unread count.
	pagedResult, err := regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Status: miniflux.EntryStatusUnread, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	if len(pagedResult.EntryIDs) != 1 {
		t.Fatalf(`Expected 1 entry ID with limit=1, got %d`, len(pagedResult.EntryIDs))
	}

	if pagedResult.Total != allEntries.Total-1 {
		t.Fatalf(`Expected total %d with limit=1, got %d`, allEntries.Total-1, pagedResult.Total)
	}

	// offset=1 should skip the first entry.
	offsetResult, err := regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Status: miniflux.EntryStatusUnread, Limit: 1, Offset: 1})
	if err != nil {
		t.Fatal(err)
	}

	if len(offsetResult.EntryIDs) != 1 {
		t.Fatalf(`Expected 1 entry ID with limit=1 offset=1, got %d`, len(offsetResult.EntryIDs))
	}

	if offsetResult.EntryIDs[0] == pagedResult.EntryIDs[0] {
		t.Fatalf(`Entry at offset=1 should differ from offset=0, both returned %d`, offsetResult.EntryIDs[0])
	}

	// Filter by starred=true: initially no starred entries.
	starredResult, err := regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Starred: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}

	if len(starredResult.EntryIDs) != 0 {
		t.Fatalf(`Expected no starred entry IDs for a new user, got %d`, len(starredResult.EntryIDs))
	}

	// Star the first entry and verify it appears in starred results.
	if err := regularUserClient.ToggleStarred(firstEntryID); err != nil {
		t.Fatal(err)
	}

	starredResult, err = regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Starred: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}

	if len(starredResult.EntryIDs) != 1 {
		t.Fatalf(`Expected 1 starred entry ID, got %d`, len(starredResult.EntryIDs))
	}

	if starredResult.Total != 1 {
		t.Fatalf(`Expected total 1, got %d`, starredResult.Total)
	}

	if starredResult.EntryIDs[0] != firstEntryID {
		t.Fatalf(`Expected starred entry ID %d, got %d`, firstEntryID, starredResult.EntryIDs[0])
	}

	// The read starred entry should appear when filtering by starred=true (read status does not affect it).
	starredResult, err = regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Starred: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}

	if len(starredResult.EntryIDs) != 1 {
		t.Fatalf(`Expected starred entry ID to persist after marking as read, got %d result(s)`, len(starredResult.EntryIDs))
	}

	// starred=false should exclude the starred entry.
	notStarredResult, err := regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Starred: boolPtr(false)})
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range notStarredResult.EntryIDs {
		if id == firstEntryID {
			t.Fatalf(`Starred entry %d should not appear in starred=false results`, firstEntryID)
		}
	}

	// Pagination with offset past the single starred result should return 0 entries but total 1.
	pagedStarred, err := regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Starred: boolPtr(true), Limit: 0, Offset: 1})
	if err != nil {
		t.Fatal(err)
	}

	if len(pagedStarred.EntryIDs) != 0 {
		t.Fatalf(`Expected 0 entry IDs with offset=1 past the only result, got %d`, len(pagedStarred.EntryIDs))
	}

	if pagedStarred.Total != 1 {
		t.Fatalf(`Expected total 1 with offset past results, got %d`, pagedStarred.Total)
	}

	// Unstarring the entry should remove it from starred=true results.
	if err := regularUserClient.ToggleStarred(firstEntryID); err != nil {
		t.Fatal(err)
	}

	starredResult, err = regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Starred: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}

	if len(starredResult.EntryIDs) != 0 {
		t.Fatalf(`Expected no starred entry IDs after unstarring, got %d`, len(starredResult.EntryIDs))
	}

	if starredResult.Total != 0 {
		t.Fatalf(`Expected total 0 after unstarring, got %d`, starredResult.Total)
	}

	// Invalid starred value should return 400.
	_, err = regularUserClient.EntryIDs(&miniflux.EntryIDsFilter{Status: "maybe"})
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Expected error for invalid status parameter, got nil, got %v`, err)
	}
}

func TestUpdateEntryEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, nil)

	entryUpdateRequest := &miniflux.EntryModificationRequest{
		Title:   new("New title"),
		Content: new("New content"),
	}

	updatedEntry, err := regularUserClient.UpdateEntry(result.Entries[0].ID, entryUpdateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if updatedEntry.Title != "New title" {
		t.Errorf(`Invalid title, got %q`, updatedEntry.Title)
	}

	if updatedEntry.Content != "New content" {
		t.Errorf(`Invalid content, got %q`, updatedEntry.Content)
	}

	entry, err := regularUserClient.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Title != "New title" {
		t.Errorf(`Invalid title, got %q`, entry.Title)
	}

	if entry.Content != "New content" {
		t.Errorf(`Invalid content, got %q`, entry.Content)
	}
}

func TestToggleStarredEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, regularTestUser := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, &miniflux.Filter{Limit: 1})

	if err := regularUserClient.ToggleStarred(result.Entries[0].ID); err != nil {
		t.Fatal(err)
	}

	entry, err := regularUserClient.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if !entry.Starred {
		t.Fatalf(`The entry should be starred`)
	}

	// The bookmark route is an alias of the star route.
	path := fmt.Sprintf("/v1/entries/%d/bookmark", entry.ID)
	if status := testConfig.requestStatus(t, regularTestUser, http.MethodPut, path); status != http.StatusNoContent {
		t.Fatalf(`Toggling the bookmark should return status %d, got %d`, http.StatusNoContent, status)
	}

	entry, err = regularUserClient.Entry(entry.ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Starred {
		t.Fatalf(`The entry should no longer be starred`)
	}
}

func TestShareEntryEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, &miniflux.Filter{Limit: 1})

	entryID := result.Entries[0].ID

	entryBeforeShare, err := regularUserClient.Entry(entryID)
	if err != nil {
		t.Fatal(err)
	}

	shareCode, err := regularUserClient.ShareEntry(entryID)
	if err != nil {
		t.Fatal(err)
	}
	if len(shareCode) < 8 {
		t.Fatal(`The entry should have a share code`)
	}

	existingShareCode, err := regularUserClient.ShareEntry(entryID)
	if err != nil {
		t.Fatal(err)
	}
	if existingShareCode != shareCode {
		t.Fatalf(`Sharing an entry twice should return the same code, got %q and %q`, shareCode, existingShareCode)
	}

	entryAfterShare, err := regularUserClient.Entry(entryID)
	if err != nil {
		t.Fatal(err)
	}

	if entryAfterShare.ShareCode != shareCode {
		t.Fatal(`The entry share code should match`)
	}

	if !entryAfterShare.ChangedAt.After(entryBeforeShare.ChangedAt) {
		t.Fatalf(`Sharing should update the entry change date, got %v before and %v after`, entryBeforeShare.ChangedAt, entryAfterShare.ChangedAt)
	}

	if err := regularUserClient.UnshareEntry(entryID); err != nil {
		t.Fatal(err)
	}

	entryAfterUnshare, err := regularUserClient.Entry(entryID)
	if err != nil {
		t.Fatal(err)
	}

	if entryAfterUnshare.ShareCode != "" {
		t.Fatal(`The share code should be deleted`)
	}

	if !entryAfterUnshare.ChangedAt.After(entryAfterShare.ChangedAt) {
		t.Fatalf(`Unsharing should update the entry change date, got %v before and %v after`, entryAfterShare.ChangedAt, entryAfterUnshare.ChangedAt)
	}

	nonExistingEntryID := int64(math.MaxInt64)
	if _, err = regularUserClient.ShareEntry(nonExistingEntryID); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Expected Not Found when sharing non existent entry: %v`, err)
	}
	if err = regularUserClient.UnshareEntry(nonExistingEntryID); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Expected Not Found when unsharing non existent entry: %v`, err)
	}

	otherUserClient, _ := testConfig.newRegularUser(t)

	otherFeedID := fixture.Subscribe(t, otherUserClient)

	result = listFeedEntries(t, otherUserClient, otherFeedID, &miniflux.Filter{Limit: 1})

	otherUserEntryID := result.Entries[0].ID
	if _, err = regularUserClient.ShareEntry(otherUserEntryID); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Expected Not Found when sharing other user's entry: %v`, err)
	}
	if err = regularUserClient.UnshareEntry(otherUserEntryID); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Expected Not Found when unsharing other user's entry: %v`, err)
	}
}

func TestSaveEntryEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, &miniflux.Filter{Limit: 1})

	if err := regularUserClient.SaveEntry(result.Entries[0].ID); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Saving an entry should raise a bad request error because no integration is configured`)
	}
}

func TestFetchContentEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, &miniflux.Filter{Limit: 1})

	content, err := regularUserClient.FetchEntryOriginalContent(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(content, servertest.ArticleText) {
		t.Fatalf(`The content should come from the original article page, got %q`, content)
	}
}

func TestFlushHistoryEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, regularTestUser := testConfig.newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, &miniflux.Filter{Limit: 3})

	// The API replies before the history is flushed in the background.
	waitForNoReadEntries := func() {
		t.Helper()

		deadline := time.Now().Add(5 * time.Second)
		for {
			readEntries, err := regularUserClient.Entries(&miniflux.Filter{Status: miniflux.EntryStatusRead})
			if err != nil {
				t.Fatal(err)
			}
			if readEntries.Total == 0 {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf(`Expected no read entries after flushing the history, got %d`, readEntries.Total)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	if err := regularUserClient.UpdateEntries([]int64{result.Entries[0].ID, result.Entries[1].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.FlushHistory(); err != nil {
		t.Fatal(err)
	}
	waitForNoReadEntries()

	// The endpoint also accepts the DELETE method, which the client does not use.
	if err := regularUserClient.UpdateEntries([]int64{result.Entries[2].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}

	if status := testConfig.requestStatus(t, regularTestUser, http.MethodDelete, "/v1/flush-history"); status != http.StatusAccepted {
		t.Fatalf(`Flushing the history with DELETE should return status %d, got %d`, http.StatusAccepted, status)
	}
	waitForNoReadEntries()
}

func TestImportFeedEntryEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	client := testConfig.newAdminClient()

	// Create a feed
	feedID := fixture.Subscribe(t, client)
	defer client.DeleteFeed(feedID)

	payload := map[string]any{
		"title":        "Imported Entry",
		"url":          "https://example.org/imported-entry",
		"content":      "Hello world",
		"external_id":  "integration-test-entry-1",
		"status":       model.EntryStatusUnread,
		"starred":      false,
		"published_at": 0,
	}

	// First import
	firstID, err := client.ImportFeedEntry(feedID, payload)
	if err != nil {
		t.Fatal(err)
	}

	if firstID == 0 {
		t.Fatal("expected non-zero entry ID on first import")
	}

	// Second import (same payload)
	secondID, err := client.ImportFeedEntry(feedID, payload)
	if err != nil {
		t.Fatal(err)
	}

	if secondID != firstID {
		t.Fatalf("expected same entry ID on re-import, got %d and %d", firstID, secondID)
	}
}
