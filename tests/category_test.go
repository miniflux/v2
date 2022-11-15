// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

//go:build integration
// +build integration

package tests

import (
	"testing"

	miniflux "miniflux.app/client"
)

func TestCreateCategory(t *testing.T) {
	categoryName := "My category"
	client := createClient(t)
	category, err := client.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	if category.ID == 0 {
		t.Fatalf(`Invalid categoryID, got "%v"`, category.ID)
	}

	if category.UserID <= 0 {
		t.Fatalf(`Invalid userID, got "%v"`, category.UserID)
	}

	if category.Title != categoryName {
		t.Fatalf(`Invalid title, got "%v" instead of "%v"`, category.Title, categoryName)
	}
}

func TestCreateCategoryWithEmptyTitle(t *testing.T) {
	client := createClient(t)
	_, err := client.CreateCategory("")
	if err == nil {
		t.Fatal(`The category title should be mandatory`)
	}
}

func TestCannotCreateDuplicatedCategory(t *testing.T) {
	client := createClient(t)

	categoryName := "My category"
	_, err := client.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateCategory(categoryName)
	if err == nil {
		t.Fatal(`Duplicated categories should not be allowed`)
	}
}

func TestUpdateCategory(t *testing.T) {
	categoryName := "My category"
	client := createClient(t)
	category, err := client.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	categoryName = "Updated category"
	category, err = client.UpdateCategory(category.ID, categoryName)
	if err != nil {
		t.Fatal(err)
	}

	if category.ID == 0 {
		t.Fatalf(`Invalid categoryID, got "%v"`, category.ID)
	}

	if category.UserID <= 0 {
		t.Fatalf(`Invalid userID, got "%v"`, category.UserID)
	}

	if category.Title != categoryName {
		t.Fatalf(`Invalid title, got %q instead of %q`, category.Title, categoryName)
	}
}

func TestUpdateInexistingCategory(t *testing.T) {
	client := createClient(t)

	_, err := client.UpdateCategory(4200000, "Test")
	if err != miniflux.ErrNotFound {
		t.Errorf(`Updating an inexisting category should returns a 404 instead of %v`, err)
	}
}

func TestMarkCategoryAsRead(t *testing.T) {
	client := createClient(t)

	feed, category := createFeed(t, client)

	results, err := client.FeedEntries(feed.ID, nil)
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}
	if results.Total == 0 {
		t.Fatalf(`Invalid number of entries: %d`, results.Total)
	}
	if results.Entries[0].Status != miniflux.EntryStatusUnread {
		t.Fatalf(`Invalid entry status, got %q instead of %q`, results.Entries[0].Status, miniflux.EntryStatusUnread)
	}

	if err := client.MarkCategoryAsRead(category.ID); err != nil {
		t.Fatalf(`Failed to mark category as read: %v`, err)
	}

	results, err = client.FeedEntries(feed.ID, nil)
	if err != nil {
		t.Fatalf(`Failed to get updated entries: %v`, err)
	}

	for _, entry := range results.Entries {
		if entry.Status != miniflux.EntryStatusRead {
			t.Errorf(`Status for entry %d was %q instead of %q`, entry.ID, entry.Status, miniflux.EntryStatusRead)
		}
	}
}

func TestListCategories(t *testing.T) {
	categoryName := "My category"
	client := createClient(t)

	_, err := client.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	if len(categories) != 2 {
		t.Fatalf(`Invalid number of categories, got "%v" instead of "%v"`, len(categories), 2)
	}

	if categories[0].ID == 0 {
		t.Fatalf(`Invalid categoryID, got "%v"`, categories[0].ID)
	}

	if categories[0].UserID <= 0 {
		t.Fatalf(`Invalid userID, got "%v"`, categories[0].UserID)
	}

	if categories[0].Title != "All" {
		t.Fatalf(`Invalid title, got "%v" instead of "%v"`, categories[0].Title, "All")
	}

	if categories[1].ID == 0 {
		t.Fatalf(`Invalid categoryID, got "%v"`, categories[0].ID)
	}

	if categories[1].UserID <= 0 {
		t.Fatalf(`Invalid userID, got "%v"`, categories[1].UserID)
	}

	if categories[1].Title != categoryName {
		t.Fatalf(`Invalid title, got "%v" instead of "%v"`, categories[1].Title, categoryName)
	}
}

func TestDeleteCategory(t *testing.T) {
	client := createClient(t)

	category, err := client.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	err = client.DeleteCategory(category.ID)
	if err != nil {
		t.Fatal(`Removing a category should not raise any error`)
	}
}

func TestCannotDeleteCategoryOfAnotherUser(t *testing.T) {
	client := createClient(t)
	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	client = createClient(t)
	err = client.DeleteCategory(categories[0].ID)
	if err == nil {
		t.Fatal(`Removing a category that belongs to another user should be forbidden`)
	}
}
