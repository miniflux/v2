// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestCreateCategoryEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	categoryName := "My category"
	category, err := regularUserClient.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	if category.ID == 0 {
		t.Errorf(`Invalid categoryID, got "%v"`, category.ID)
	}

	if category.UserID <= 0 {
		t.Errorf(`Invalid userID, got "%v"`, category.UserID)
	}

	if category.Title != categoryName {
		t.Errorf(`Invalid title, got "%v" instead of "%v"`, category.Title, categoryName)
	}

	if category.HideGlobally {
		t.Errorf(`Invalid hide globally value, got "%v"`, category.HideGlobally)
	}
}

func TestCreateCategoryWithEmptyTitle(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := testConfig.newAdminClient()
	_, err := client.CreateCategory("")
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating a category with an empty title should raise an error, got %v`, err)
	}
}

func TestCannotCreateDuplicatedCategory(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	regularUserClient, _ := testConfig.newRegularUser(t)
	categoryName := "My category"

	if _, err := regularUserClient.CreateCategory(categoryName); err != nil {
		t.Fatal(err)
	}

	if _, err := regularUserClient.CreateCategory(categoryName); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Duplicated categories should not be allowed, got %v`, err)
	}
}

func TestCreateCategoryWithOptions(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	newCategory, err := regularUserClient.CreateCategoryWithOptions(&miniflux.CategoryCreationRequest{
		Title:        "My category",
		HideGlobally: true,
	})
	if err != nil {
		t.Fatalf(`Creating a category with options should not raise an error: %v`, err)
	}

	categories, err := regularUserClient.Categories()
	if err != nil {
		t.Fatal(err)
	}

	for _, category := range categories {
		if category.ID == newCategory.ID {
			if category.Title != newCategory.Title {
				t.Errorf(`Invalid title, got %q instead of %q`, category.Title, newCategory.Title)
			}
			if category.HideGlobally != true {
				t.Errorf(`Invalid hide globally value, got "%v"`, category.HideGlobally)
			}
			break
		}
	}
}

func TestUpdateCategoryEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	regularUserClient, regularTestUser := testConfig.newRegularUser(t)

	categoryName := "My category"
	category, err := regularUserClient.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	updatedCategory, err := regularUserClient.UpdateCategory(category.ID, "new title")
	if err != nil {
		t.Fatal(err)
	}

	if updatedCategory.ID != category.ID {
		t.Errorf(`Invalid categoryID, got "%v"`, updatedCategory.ID)
	}

	if updatedCategory.UserID != regularTestUser.ID {
		t.Errorf(`Invalid userID, got "%v"`, updatedCategory.UserID)
	}

	if updatedCategory.Title != "new title" {
		t.Errorf(`Invalid title, got "%v" instead of "%v"`, updatedCategory.Title, "new title")
	}

	if updatedCategory.HideGlobally {
		t.Errorf(`Invalid hide globally value, got "%v"`, updatedCategory.HideGlobally)
	}
}

func TestUpdateCategoryWithOptions(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	newCategory, err := regularUserClient.CreateCategoryWithOptions(&miniflux.CategoryCreationRequest{
		Title: "My category",
	})
	if err != nil {
		t.Fatalf(`Creating a category with options should not raise an error: %v`, err)
	}

	updatedCategory, err := regularUserClient.UpdateCategoryWithOptions(newCategory.ID, &miniflux.CategoryModificationRequest{
		Title: new("new title"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if updatedCategory.ID != newCategory.ID {
		t.Errorf(`Invalid categoryID, got "%v"`, updatedCategory.ID)
	}

	if updatedCategory.Title != "new title" {
		t.Errorf(`Invalid title, got "%v" instead of "%v"`, updatedCategory.Title, "new title")
	}

	if updatedCategory.HideGlobally {
		t.Errorf(`Invalid hide globally value, got "%v"`, updatedCategory.HideGlobally)
	}

	updatedCategory, err = regularUserClient.UpdateCategoryWithOptions(newCategory.ID, &miniflux.CategoryModificationRequest{
		HideGlobally: new(true),
	})
	if err != nil {
		t.Fatal(err)
	}

	if updatedCategory.ID != newCategory.ID {
		t.Errorf(`Invalid categoryID, got "%v"`, updatedCategory.ID)
	}

	if updatedCategory.Title != "new title" {
		t.Errorf(`Invalid title, got "%v" instead of "%v"`, updatedCategory.Title, "new title")
	}

	if !updatedCategory.HideGlobally {
		t.Errorf(`Invalid hide globally value, got "%v"`, updatedCategory.HideGlobally)
	}

	updatedCategory, err = regularUserClient.UpdateCategoryWithOptions(newCategory.ID, &miniflux.CategoryModificationRequest{
		HideGlobally: new(false),
	})
	if err != nil {
		t.Fatal(err)
	}

	if updatedCategory.ID != newCategory.ID {
		t.Errorf(`Invalid categoryID, got %d`, updatedCategory.ID)
	}

	if updatedCategory.Title != "new title" {
		t.Errorf(`Invalid title, got %q instead of %q`, updatedCategory.Title, "new title")
	}

	if updatedCategory.HideGlobally {
		t.Errorf(`Invalid hide globally value, got "%v"`, updatedCategory.HideGlobally)
	}
}

func TestUpdateInexistingCategory(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := testConfig.newAdminClient()
	_, err := client.UpdateCategory(123456789, "new title")
	if !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Updating an inexisting category should raise an error, got %v`, err)
	}
}

func TestDeleteCategoryEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	categoryName := "My category"
	category, err := regularUserClient.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.DeleteCategory(category.ID); err != nil {
		t.Fatal(err)
	}
}

func TestCannotDeleteInexistingCategory(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := testConfig.newAdminClient()
	err := client.DeleteCategory(123456789)
	if !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Deleting an inexisting category should raise an error, got %v`, err)
	}
}

func TestCannotDeleteCategoryOfAnotherUser(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	adminClient := testConfig.newAdminClient()
	regularUserClient, _ := testConfig.newRegularUser(t)

	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	err = adminClient.DeleteCategory(category.ID)
	if !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Regular users should not be able to delete categories of other users, got %v`, err)
	}
}

func TestGetCategoriesEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	regularUserClient, regularTestUser := testConfig.newRegularUser(t)

	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	categories, err := regularUserClient.Categories()
	if err != nil {
		t.Fatal(err)
	}

	if len(categories) != 2 {
		t.Fatalf(`Invalid number of categories, got %d instead of %d`, len(categories), 1)
	}

	if categories[0].UserID != regularTestUser.ID {
		t.Fatalf(`Invalid userID, got %d`, categories[0].UserID)
	}

	if categories[0].Title != "All" {
		t.Fatalf(`Invalid title, got %q instead of %q`, categories[0].Title, "All")
	}

	if categories[0].FeedCount != nil {
		t.Errorf(`Expected FeedCount to be nil, got %d`, *categories[0].FeedCount)
	}

	if categories[0].TotalUnread != nil {
		t.Errorf(`Expected TotalUnread to be nil, got %d`, *categories[0].TotalUnread)
	}

	if categories[1].ID != category.ID {
		t.Fatalf(`Invalid categoryID, got %d`, categories[0].ID)
	}

	if categories[1].UserID != regularTestUser.ID {
		t.Fatalf(`Invalid userID, got %d`, categories[0].UserID)
	}

	if categories[1].Title != "My category" {
		t.Fatalf(`Invalid title, got %q instead of %q`, categories[0].Title, "My category")
	}

	if categories[1].FeedCount != nil {
		t.Errorf(`Expected FeedCount to be nil, got %d`, *categories[1].FeedCount)
	}

	if categories[1].TotalUnread != nil {
		t.Errorf(`Expected TotalUnread to be nil, got %d`, *categories[1].TotalUnread)
	}

	categories, err = regularUserClient.CategoriesWithCounters()
	if err != nil {
		t.Fatal(err)
	}

	if len(categories) != 2 {
		t.Fatalf(`Invalid number of categories, got %d instead of %d`, len(categories), 1)
	}

	if categories[1].FeedCount == nil {
		t.Fatalf(`Expected FeedCount to be not nil`)
	}

	if categories[1].TotalUnread == nil {
		t.Fatalf(`Expected TotalUnread to be not nil`)
	}

	expectedCounterValue := 0
	if *categories[1].FeedCount != expectedCounterValue {
		t.Errorf(`Expected FeedCount to be %d, got %d`, expectedCounterValue, *categories[1].FeedCount)
	}

	if *categories[1].TotalUnread != expectedCounterValue {
		t.Errorf(`Expected TotalUnread to be %d, got %d`, expectedCounterValue, *categories[1].TotalUnread)
	}
}

func TestMarkCategoryAsReadEndpoint(t *testing.T) {
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

	if err := regularUserClient.MarkCategoryAsRead(category.ID); err != nil {
		t.Fatal(err)
	}

	results := listFeedEntries(t, regularUserClient, feedID, nil)

	for _, entry := range results.Entries {
		if entry.Status != miniflux.EntryStatusRead {
			t.Errorf(`Status for entry %d was %q instead of %q`, entry.ID, entry.Status, miniflux.EntryStatusRead)
		}
	}
}

func TestRefreshCategoryEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feed, err := regularUserClient.Feed(fixture.Subscribe(t, regularUserClient))
	if err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.RefreshCategory(feed.Category.ID); err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.RefreshCategory(0); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Refreshing a category with an invalid ID should raise a bad request error, got %v`, err)
	}

	if err := regularUserClient.RefreshCategory(123456789); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Refreshing an inexisting category should raise a "not found" error, got %v`, err)
	}

	otherUserClient, _ := testConfig.newRegularUser(t)
	otherCategory, err := otherUserClient.CreateCategory("Other category")
	if err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.RefreshCategory(otherCategory.ID); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Refreshing a category of another user should raise a "not found" error, got %v`, err)
	}
}
