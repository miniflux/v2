// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package tests

import (
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestGetAllFeedEntries(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	allResults, err := client.FeedEntries(feed.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	if allResults.Total == 0 {
		t.Fatal(`Invalid number of entries`)
	}

	if allResults.Entries[0].Title == "" {
		t.Fatal(`Invalid entry title`)
	}

	filteredResults, err := client.FeedEntries(feed.ID, &miniflux.Filter{Limit: 1, Offset: 5})
	if err != nil {
		t.Fatal(err)
	}

	if allResults.Total != filteredResults.Total {
		t.Fatal(`Total should always contains the total number of items regardless of filters`)
	}

	if allResults.Entries[0].ID == filteredResults.Entries[0].ID {
		t.Fatal(`Filtered entries should be different than previous results`)
	}

	filteredResultsByEntryID, err := client.FeedEntries(feed.ID, &miniflux.Filter{AfterEntryID: allResults.Entries[0].ID})
	if err != nil {
		t.Fatal(err)
	}

	if filteredResultsByEntryID.Entries[0].ID == allResults.Entries[0].ID {
		t.Fatal(`The first entry should be filtered out`)
	}
}

func TestGetAllCategoryEntries(t *testing.T) {
	client := createClient(t)
	_, category := createFeed(t, client)

	allResults, err := client.CategoryEntries(category.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	if allResults.Total == 0 {
		t.Fatal(`Invalid number of entries`)
	}

	if allResults.Entries[0].Title == "" {
		t.Fatal(`Invalid entry title`)
	}

	filteredResults, err := client.CategoryEntries(category.ID, &miniflux.Filter{Limit: 1, Offset: 5})
	if err != nil {
		t.Fatal(err)
	}

	if allResults.Total != filteredResults.Total {
		t.Fatal(`Total should always contains the total number of items regardless of filters`)
	}

	if allResults.Entries[0].ID == filteredResults.Entries[0].ID {
		t.Fatal(`Filtered entries should be different than previous results`)
	}

	filteredResultsByEntryID, err := client.CategoryEntries(category.ID, &miniflux.Filter{AfterEntryID: allResults.Entries[0].ID})
	if err != nil {
		t.Fatal(err)
	}

	if filteredResultsByEntryID.Entries[0].ID == allResults.Entries[0].ID {
		t.Fatal(`The first entry should be filtered out`)
	}
}

func TestGetAllEntries(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	resultWithoutSorting, err := client.Entries(nil)
	if err != nil {
		t.Fatal(err)
	}

	if resultWithoutSorting.Total == 0 {
		t.Fatal(`Invalid number of entries`)
	}

	resultWithStatusFilter, err := client.Entries(&miniflux.Filter{Status: miniflux.EntryStatusRead})
	if err != nil {
		t.Fatal(err)
	}

	if resultWithStatusFilter.Total != 0 {
		t.Fatal(`We should have 0 read entries`)
	}

	resultWithDifferentSorting, err := client.Entries(&miniflux.Filter{Order: "published_at", Direction: "desc"})
	if err != nil {
		t.Fatal(err)
	}

	if resultWithDifferentSorting.Entries[0].Title == resultWithoutSorting.Entries[0].Title {
		t.Fatalf(`The items should be sorted differently "%v" vs "%v"`, resultWithDifferentSorting.Entries[0].Title, resultWithoutSorting.Entries[0].Title)
	}

	resultWithStarredEntries, err := client.Entries(&miniflux.Filter{Starred: miniflux.FilterOnlyStarred})
	if err != nil {
		t.Fatal(err)
	}

	if resultWithStarredEntries.Total != 0 {
		t.Fatalf(`We are not supposed to have starred entries yet`)
	}
}

func TestFilterEntriesByCategory(t *testing.T) {
	client := createClient(t)
	category, err := client.CreateCategory("Test Filter by Category")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testFeedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	results, err := client.Entries(&miniflux.Filter{CategoryID: category.ID})
	if err != nil {
		t.Fatal(err)
	}

	if results.Total == 0 {
		t.Fatalf(`We should have more than one entry`)
	}

	if results.Entries[0].Feed.Category == nil {
		t.Fatalf(`The entry feed category should not be nil`)
	}

	if results.Entries[0].Feed.Category.ID != category.ID {
		t.Errorf(`Entries should be filtered by category_id=%d`, category.ID)
	}
}

func TestFilterEntriesByFeed(t *testing.T) {
	client := createClient(t)
	category, err := client.CreateCategory("Test Filter by Feed")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testFeedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	results, err := client.Entries(&miniflux.Filter{FeedID: feedID})
	if err != nil {
		t.Fatal(err)
	}

	if results.Total == 0 {
		t.Fatalf(`We should have more than one entry`)
	}

	if results.Entries[0].Feed.Category == nil {
		t.Fatalf(`The entry feed category should not be nil`)
	}

	if results.Entries[0].Feed.Category.ID != category.ID {
		t.Errorf(`Entries should be filtered by category_id=%d`, category.ID)
	}
}

func TestFilterEntriesByStatuses(t *testing.T) {
	client := createClient(t)
	category, err := client.CreateCategory("Test Filter by statuses")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testFeedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	results, err := client.Entries(&miniflux.Filter{FeedID: feedID})
	if err != nil {
		t.Fatal(err)
	}

	if err := client.UpdateEntries([]int64{results.Entries[0].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}

	if err := client.UpdateEntries([]int64{results.Entries[1].ID}, miniflux.EntryStatusRemoved); err != nil {
		t.Fatal(err)
	}

	results, err = client.Entries(&miniflux.Filter{Statuses: []string{miniflux.EntryStatusRead, miniflux.EntryStatusRemoved}})
	if err != nil {
		t.Fatal(err)
	}

	if results.Total != 2 {
		t.Fatalf(`We should have 2 entries`)
	}

	if results.Entries[0].Status != "read" {
		t.Errorf(`The first entry has the wrong status: %s`, results.Entries[0].Status)
	}

	if results.Entries[1].Status != "removed" {
		t.Errorf(`The 2nd entry has the wrong status: %s`, results.Entries[1].Status)
	}
}

func TestSearchEntries(t *testing.T) {
	client := createClient(t)
	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testFeedURL,
		CategoryID: categories[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	results, err := client.Entries(&miniflux.Filter{Search: "2.0.8"})
	if err != nil {
		t.Fatal(err)
	}

	if results.Total != 1 {
		t.Fatalf(`We should have only one entry instead of %d`, results.Total)
	}
}

func TestInvalidFilters(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	_, err := client.Entries(&miniflux.Filter{Status: "invalid"})
	if err == nil {
		t.Fatal(`Using invalid status should raise an error`)
	}

	_, err = client.Entries(&miniflux.Filter{Direction: "invalid"})
	if err == nil {
		t.Fatal(`Using invalid direction should raise an error`)
	}

	_, err = client.Entries(&miniflux.Filter{Order: "invalid"})
	if err == nil {
		t.Fatal(`Using invalid order should raise an error`)
	}
}

func TestGetFeedEntry(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	result, err := client.Entries(&miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	// Test get entry by entry id and feed id
	entry, err := client.FeedEntry(result.Entries[0].FeedID, result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if entry.ID != result.Entries[0].ID {
		t.Fatal("Wrong entry returned")
	}
}

func TestGetCategoryEntry(t *testing.T) {
	client := createClient(t)
	_, category := createFeed(t, client)

	result, err := client.Entries(&miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	// Test get entry by entry id and category id
	entry, err := client.CategoryEntry(category.ID, result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if entry.ID != result.Entries[0].ID {
		t.Fatal("Wrong entry returned")
	}
}

func TestGetEntry(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	result, err := client.Entries(&miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	// Test get entry by entry id only
	entry, err := client.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if entry.ID != result.Entries[0].ID {
		t.Fatal("Wrong entry returned")
	}
}

func TestUpdateStatus(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	result, err := client.Entries(&miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	err = client.UpdateEntries([]int64{result.Entries[0].ID}, miniflux.EntryStatusRead)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := client.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Status != miniflux.EntryStatusRead {
		t.Fatal("The entry status should be updated")
	}

	err = client.UpdateEntries([]int64{result.Entries[0].ID}, "invalid")
	if err == nil {
		t.Fatal(`Invalid entry status should not be accepted`)
	}

	err = client.UpdateEntries([]int64{}, miniflux.EntryStatusRead)
	if err == nil {
		t.Fatal(`An empty list of entry should not be accepted`)
	}
}

func TestUpdateEntry(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	result, err := client.Entries(&miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	title := "New title"
	content := "New content"

	_, err = client.UpdateEntry(result.Entries[0].ID, &miniflux.EntryModificationRequest{
		Title:   &title,
		Content: &content,
	})
	if err != nil {
		t.Fatal(err)
	}

	entry, err := client.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Title != title {
		t.Fatal("The entry title should be updated")
	}

	if entry.Content != content {
		t.Fatal("The entry content should be updated")
	}
}

func TestToggleBookmark(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	result, err := client.Entries(&miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	if result.Entries[0].Starred {
		t.Fatal("The entry should not be starred")
	}

	err = client.ToggleBookmark(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := client.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if !entry.Starred {
		t.Fatal("The entry should be starred")
	}
}

func TestHistoryOrder(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	result, err := client.Entries(&miniflux.Filter{Limit: 3})
	if err != nil {
		t.Fatal(err)
	}

	selectedEntryID := result.Entries[2].ID

	err = client.UpdateEntries([]int64{selectedEntryID}, miniflux.EntryStatusRead)
	if err != nil {
		t.Fatal(err)
	}

	history, err := client.Entries(&miniflux.Filter{Order: "changed_at", Direction: "desc", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	if history.Entries[0].ID != selectedEntryID {
		t.Fatal("The entry that we just read should be at the top of the history")
	}
}

func TestFlushHistory(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	result, err := client.Entries(&miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	selectedEntryID := result.Entries[0].ID

	err = client.UpdateEntries([]int64{selectedEntryID}, miniflux.EntryStatusRead)
	if err != nil {
		t.Fatal(err)
	}

	err = client.FlushHistory()
	if err != nil {
		t.Fatal(err)
	}

	history, err := client.Entries(&miniflux.Filter{Status: miniflux.EntryStatusRemoved})
	if err != nil {
		t.Fatal(err)
	}

	if history.Entries[0].ID != selectedEntryID {
		t.Fatal("The entry that we just read should have the removed status")
	}
}
