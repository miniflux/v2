// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// +build integration

package tests

import (
	"testing"

	miniflux "miniflux.app/client"
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

	filteredResultsByEntryID, err := client.FeedEntries(feed.ID, &miniflux.Filter{BeforeEntryID: allResults.Entries[0].ID})
	if err != nil {
		t.Fatal(err)
	}

	if filteredResultsByEntryID.Entries[0].ID == allResults.Entries[0].ID {
		t.Fatal(`The first entry should filtered out`)
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

	resultWithStarredEntries, err := client.Entries(&miniflux.Filter{Starred: true})
	if err != nil {
		t.Fatal(err)
	}

	if resultWithStarredEntries.Total != 0 {
		t.Fatalf(`We are not supposed to have starred entries yet`)
	}
}

func TestSearchEntries(t *testing.T) {
	client := createClient(t)
	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed("https://github.com/miniflux/miniflux/releases.atom", categories[0].ID)
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

func TestGetEntry(t *testing.T) {
	client := createClient(t)
	createFeed(t, client)

	result, err := client.Entries(&miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}

	entry, err := client.FeedEntry(result.Entries[0].FeedID, result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.ID != result.Entries[0].ID {
		t.Fatal("Wrong entry returned")
	}

	entry, err = client.Entry(result.Entries[0].ID)
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
		t.Fatal(`Invalid entry status should ne be accepted`)
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
