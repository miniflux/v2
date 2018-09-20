// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// +build integration

package tests

import (
	"strings"
	"testing"

	miniflux "miniflux.app/client"
)

func TestCreateFeed(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	if feed.ID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feed.ID)
	}
}

func TestCannotCreateDuplicatedFeed(t *testing.T) {
	client := createClient(t)
	feed, category := createFeed(t, client)

	_, err := client.CreateFeed(feed.FeedURL, category.ID)
	if err == nil {
		t.Fatal(`Duplicated feeds should not be allowed`)
	}
}

func TestCreateFeedWithInexistingCategory(t *testing.T) {
	client := createClient(t)

	_, err := client.CreateFeed(testFeedURL, -1)
	if err == nil {
		t.Fatal(`Feeds should not be created with inexisting category`)
	}
}

func TestUpdateFeedURL(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	url := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{FeedURL: &url})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.FeedURL != url {
		t.Fatalf(`Wrong FeedURL, got %q instead of %q`, updatedFeed.FeedURL, url)
	}

	url = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{FeedURL: &url})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.FeedURL == "" {
		t.Fatalf(`The FeedURL should not be empty`)
	}
}

func TestUpdateFeedSiteURL(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	url := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{SiteURL: &url})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.SiteURL != url {
		t.Fatalf(`Wrong SiteURL, got %q instead of %q`, updatedFeed.SiteURL, url)
	}

	url = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{SiteURL: &url})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.SiteURL == "" {
		t.Fatalf(`The SiteURL should not be empty`)
	}
}

func TestUpdateFeedTitle(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	newTitle := "My new feed"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{Title: &newTitle})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Title != newTitle {
		t.Fatalf(`Wrong title, got %q instead of %q`, updatedFeed.Title, newTitle)
	}

	newTitle = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{Title: &newTitle})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Title == "" {
		t.Fatalf(`The Title should not be empty`)
	}
}

func TestUpdateFeedCrawler(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	crawler := true
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{Crawler: &crawler})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Crawler != crawler {
		t.Fatalf(`Wrong crawler value, got "%v" instead of "%v"`, updatedFeed.Crawler, crawler)
	}

	if updatedFeed.Title != feed.Title {
		t.Fatalf(`The titles should be the same after update`)
	}

	crawler = false
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{Crawler: &crawler})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Crawler != crawler {
		t.Fatalf(`Wrong crawler value, got "%v" instead of "%v"`, updatedFeed.Crawler, crawler)
	}
}

func TestUpdateFeedScraperRules(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	scraperRules := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{ScraperRules: &scraperRules})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.ScraperRules != scraperRules {
		t.Fatalf(`Wrong ScraperRules value, got "%v" instead of "%v"`, updatedFeed.ScraperRules, scraperRules)
	}

	scraperRules = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{ScraperRules: &scraperRules})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.ScraperRules != scraperRules {
		t.Fatalf(`Wrong ScraperRules value, got "%v" instead of "%v"`, updatedFeed.ScraperRules, scraperRules)
	}
}

func TestUpdateFeedRewriteRules(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	rewriteRules := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{RewriteRules: &rewriteRules})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.RewriteRules != rewriteRules {
		t.Fatalf(`Wrong RewriteRules value, got "%v" instead of "%v"`, updatedFeed.RewriteRules, rewriteRules)
	}

	rewriteRules = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{RewriteRules: &rewriteRules})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.RewriteRules != rewriteRules {
		t.Fatalf(`Wrong RewriteRules value, got "%v" instead of "%v"`, updatedFeed.RewriteRules, rewriteRules)
	}
}

func TestUpdateFeedUserAgent(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	userAgent := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{UserAgent: &userAgent})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.UserAgent != userAgent {
		t.Fatalf(`Wrong UserAgent value, got "%v" instead of "%v"`, updatedFeed.UserAgent, userAgent)
	}

	userAgent = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{UserAgent: &userAgent})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.UserAgent != userAgent {
		t.Fatalf(`Wrong UserAgent value, got "%v" instead of "%v"`, updatedFeed.UserAgent, userAgent)
	}
}

func TestUpdateFeedUsername(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	username := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{Username: &username})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Username != username {
		t.Fatalf(`Wrong Username value, got "%v" instead of "%v"`, updatedFeed.Username, username)
	}

	username = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{Username: &username})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Username != username {
		t.Fatalf(`Wrong Username value, got "%v" instead of "%v"`, updatedFeed.Username, username)
	}
}

func TestUpdateFeedPassword(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	password := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{Password: &password})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Password != password {
		t.Fatalf(`Wrong Password value, got "%v" instead of "%v"`, updatedFeed.Password, password)
	}

	password = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{Password: &password})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Password != password {
		t.Fatalf(`Wrong Password value, got "%v" instead of "%v"`, updatedFeed.Password, password)
	}
}

func TestUpdateFeedCategory(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	newCategory, err := client.CreateCategory("my new category")
	if err != nil {
		t.Fatal(err)
	}

	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModification{CategoryID: &newCategory.ID})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Category.ID != newCategory.ID {
		t.Fatalf(`Wrong CategoryID value, got "%v" instead of "%v"`, updatedFeed.Category.ID, newCategory.ID)
	}

	categoryID := int64(0)
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModification{CategoryID: &categoryID})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Category.ID == 0 {
		t.Fatalf(`The CategoryID must defined`)
	}
}

func TestDeleteFeed(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)
	if err := client.DeleteFeed(feed.ID); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshFeed(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)
	if err := client.RefreshFeed(feed.ID); err != nil {
		t.Fatal(err)
	}
}

func TestGetFeed(t *testing.T) {
	client := createClient(t)
	feed, category := createFeed(t, client)

	if feed.Title != testFeedTitle {
		t.Fatalf(`Invalid feed title, got "%v" instead of "%v"`, feed.Title, testFeedTitle)
	}

	if feed.SiteURL != testWebsiteURL {
		t.Fatalf(`Invalid site URL, got "%v" instead of "%v"`, feed.SiteURL, testWebsiteURL)
	}

	if feed.FeedURL != testFeedURL {
		t.Fatalf(`Invalid feed URL, got "%v" instead of "%v"`, feed.FeedURL, testFeedURL)
	}

	if feed.Category.ID != category.ID {
		t.Fatalf(`Invalid feed category ID, got "%v" instead of "%v"`, feed.Category.ID, category.ID)
	}

	if feed.Category.UserID != category.UserID {
		t.Fatalf(`Invalid feed category user ID, got "%v" instead of "%v"`, feed.Category.UserID, category.UserID)
	}

	if feed.Category.Title != category.Title {
		t.Fatalf(`Invalid feed category title, got "%v" instead of "%v"`, feed.Category.Title, category.Title)
	}
}

func TestGetFeedIcon(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)
	feedIcon, err := client.FeedIcon(feed.ID)
	if err != nil {
		t.Fatal(err)
	}

	if feedIcon.ID == 0 {
		t.Fatalf(`Invalid feed icon ID, got "%v"`, feedIcon.ID)
	}

	if feedIcon.MimeType != "image/x-icon" {
		t.Fatalf(`Invalid feed icon mime type, got "%v" instead of "%v"`, feedIcon.MimeType, "image/x-icon")
	}

	if !strings.Contains(feedIcon.Data, "image/x-icon") {
		t.Fatalf(`Invalid feed icon data, got "%v"`, feedIcon.Data)
	}
}

func TestGetFeedIconNotFound(t *testing.T) {
	client := createClient(t)
	if _, err := client.FeedIcon(42); err == nil {
		t.Fatalf(`The feed icon should be null`)
	}
}

func TestGetFeeds(t *testing.T) {
	client := createClient(t)
	feed, category := createFeed(t, client)

	feeds, err := client.Feeds()
	if err != nil {
		t.Fatal(err)
	}

	if len(feeds) != 1 {
		t.Fatalf(`Invalid number of feeds`)
	}

	if feeds[0].ID != feed.ID {
		t.Fatalf(`Invalid feed ID, got "%v" instead of "%v"`, feeds[0].ID, feed.ID)
	}

	if feeds[0].Title != testFeedTitle {
		t.Fatalf(`Invalid feed title, got "%v" instead of "%v"`, feeds[0].Title, testFeedTitle)
	}

	if feeds[0].SiteURL != testWebsiteURL {
		t.Fatalf(`Invalid site URL, got "%v" instead of "%v"`, feeds[0].SiteURL, testWebsiteURL)
	}

	if feeds[0].FeedURL != testFeedURL {
		t.Fatalf(`Invalid feed URL, got "%v" instead of "%v"`, feeds[0].FeedURL, testFeedURL)
	}

	if feeds[0].Category.ID != category.ID {
		t.Fatalf(`Invalid feed category ID, got "%v" instead of "%v"`, feeds[0].Category.ID, category.ID)
	}

	if feeds[0].Category.UserID != category.UserID {
		t.Fatalf(`Invalid feed category user ID, got "%v" instead of "%v"`, feeds[0].Category.UserID, category.UserID)
	}

	if feeds[0].Category.Title != category.Title {
		t.Fatalf(`Invalid feed category title, got "%v" instead of "%v"`, feeds[0].Category.Title, category.Title)
	}
}
