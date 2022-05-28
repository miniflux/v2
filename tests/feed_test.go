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

	_, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    feed.FeedURL,
		CategoryID: category.ID,
	})
	if err == nil {
		t.Fatal(`Duplicated feeds should not be allowed`)
	}
}

func TestCreateFeedWithInexistingCategory(t *testing.T) {
	client := createClient(t)
	_, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testFeedURL,
		CategoryID: -1,
	})
	if err == nil {
		t.Fatal(`Feeds should not be created with inexisting category`)
	}
}

func TestCreateFeedWithEmptyFeedURL(t *testing.T) {
	client := createClient(t)
	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    "",
		CategoryID: categories[0].ID,
	})
	if err == nil {
		t.Fatal(`Feeds should not be created with an empty feed URL`)
	}
}

func TestCreateFeedWithInvalidFeedURL(t *testing.T) {
	client := createClient(t)
	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    "invalid",
		CategoryID: categories[0].ID,
	})
	if err == nil {
		t.Fatal(`Feeds should not be created with an invalid feed URL`)
	}
}

func TestCreateDisabledFeed(t *testing.T) {
	client := createClient(t)

	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testFeedURL,
		CategoryID: categories[0].ID,
		Disabled:   true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	feed, err := client.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if !feed.Disabled {
		t.Error(`The feed should be disabled`)
	}
}

func TestCreateFeedWithDisabledCache(t *testing.T) {
	client := createClient(t)

	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:         testFeedURL,
		CategoryID:      categories[0].ID,
		IgnoreHTTPCache: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	feed, err := client.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if !feed.IgnoreHTTPCache {
		t.Error(`The feed should be ignoring HTTP cache`)
	}
}

func TestCreateFeedWithCrawlerEnabled(t *testing.T) {
	client := createClient(t)

	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testFeedURL,
		CategoryID: categories[0].ID,
		Crawler:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	feed, err := client.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if !feed.Crawler {
		t.Error(`The feed should have the scraper enabled`)
	}
}

func TestCreateFeedWithSelfSignedCertificatesAllowed(t *testing.T) {
	client := createClient(t)

	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:                     testFeedURL,
		CategoryID:                  categories[0].ID,
		AllowSelfSignedCertificates: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	feed, err := client.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if !feed.AllowSelfSignedCertificates {
		t.Error(`The feed should have self-signed certificates enabled`)
	}
}

func TestCreateFeedWithScraperRule(t *testing.T) {
	client := createClient(t)

	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:      testFeedURL,
		CategoryID:   categories[0].ID,
		ScraperRules: "article",
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	feed, err := client.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if feed.ScraperRules != "article" {
		t.Error(`The feed should have the custom scraper rule saved`)
	}
}

func TestCreateFeedWithKeeplistRule(t *testing.T) {
	client := createClient(t)

	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:       testFeedURL,
		CategoryID:    categories[0].ID,
		KeeplistRules: "(?i)miniflux",
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got %q`, feedID)
	}

	feed, err := client.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if feed.KeeplistRules != "(?i)miniflux" {
		t.Error(`The feed should have the custom keep list rule saved`)
	}
}

func TestCreateFeedWithInvalidBlocklistRule(t *testing.T) {
	client := createClient(t)

	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:        testFeedURL,
		CategoryID:     categories[0].ID,
		BlocklistRules: "[",
	})
	if err == nil {
		t.Fatal(`Feed with invalid block list rule should not be created`)
	}
}

func TestUpdateFeedURL(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	url := "https://www.example.org/feed.xml"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{FeedURL: &url})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.FeedURL != url {
		t.Fatalf(`Wrong FeedURL, got %q instead of %q`, updatedFeed.FeedURL, url)
	}
}

func TestUpdateFeedWithEmptyFeedURL(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	url := ""
	if _, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{FeedURL: &url}); err == nil {
		t.Error(`Updating a feed with an empty feed URL should not be possible`)
	}
}

func TestUpdateFeedWithInvalidFeedURL(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	url := "invalid"
	if _, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{FeedURL: &url}); err == nil {
		t.Error(`Updating a feed with an invalid feed URL should not be possible`)
	}
}

func TestUpdateFeedSiteURL(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	url := "https://www.example.org/"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{SiteURL: &url})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.SiteURL != url {
		t.Fatalf(`Wrong SiteURL, got %q instead of %q`, updatedFeed.SiteURL, url)
	}
}

func TestUpdateFeedWithEmptySiteURL(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	url := ""
	if _, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{SiteURL: &url}); err == nil {
		t.Error(`Updating a feed with an empty site URL should not be possible`)
	}
}

func TestUpdateFeedWithInvalidSiteURL(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	url := "invalid"
	if _, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{SiteURL: &url}); err == nil {
		t.Error(`Updating a feed with an invalid site URL should not be possible`)
	}
}

func TestUpdateFeedTitle(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	newTitle := "My new feed"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Title: &newTitle})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Title != newTitle {
		t.Fatalf(`Wrong title, got %q instead of %q`, updatedFeed.Title, newTitle)
	}
}

func TestUpdateFeedWithEmptyTitle(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	title := ""
	if _, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Title: &title}); err == nil {
		t.Error(`Updating a feed with an empty title should not be possible`)
	}
}

func TestUpdateFeedCrawler(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	crawler := true
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Crawler: &crawler})
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
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Crawler: &crawler})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Crawler != crawler {
		t.Fatalf(`Wrong crawler value, got "%v" instead of "%v"`, updatedFeed.Crawler, crawler)
	}
}

func TestUpdateFeedAllowSelfSignedCertificates(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	selfSigned := true
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{AllowSelfSignedCertificates: &selfSigned})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.AllowSelfSignedCertificates != selfSigned {
		t.Fatalf(`Wrong AllowSelfSignedCertificates value, got "%v" instead of "%v"`, updatedFeed.AllowSelfSignedCertificates, selfSigned)
	}

	selfSigned = false
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{AllowSelfSignedCertificates: &selfSigned})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.AllowSelfSignedCertificates != selfSigned {
		t.Fatalf(`Wrong AllowSelfSignedCertificates value, got "%v" instead of "%v"`, updatedFeed.AllowSelfSignedCertificates, selfSigned)
	}
}

func TestUpdateFeedScraperRules(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	scraperRules := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{ScraperRules: &scraperRules})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.ScraperRules != scraperRules {
		t.Fatalf(`Wrong ScraperRules value, got "%v" instead of "%v"`, updatedFeed.ScraperRules, scraperRules)
	}

	scraperRules = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{ScraperRules: &scraperRules})
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
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{RewriteRules: &rewriteRules})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.RewriteRules != rewriteRules {
		t.Fatalf(`Wrong RewriteRules value, got "%v" instead of "%v"`, updatedFeed.RewriteRules, rewriteRules)
	}

	rewriteRules = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{RewriteRules: &rewriteRules})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.RewriteRules != rewriteRules {
		t.Fatalf(`Wrong RewriteRules value, got "%v" instead of "%v"`, updatedFeed.RewriteRules, rewriteRules)
	}
}

func TestUpdateFeedKeeplistRules(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	keeplistRules := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{KeeplistRules: &keeplistRules})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.KeeplistRules != keeplistRules {
		t.Fatalf(`Wrong KeeplistRules value, got "%v" instead of "%v"`, updatedFeed.KeeplistRules, keeplistRules)
	}

	keeplistRules = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{KeeplistRules: &keeplistRules})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.KeeplistRules != keeplistRules {
		t.Fatalf(`Wrong KeeplistRules value, got "%v" instead of "%v"`, updatedFeed.KeeplistRules, keeplistRules)
	}
}

func TestUpdateFeedUserAgent(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	userAgent := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{UserAgent: &userAgent})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.UserAgent != userAgent {
		t.Fatalf(`Wrong UserAgent value, got "%v" instead of "%v"`, updatedFeed.UserAgent, userAgent)
	}

	userAgent = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{UserAgent: &userAgent})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.UserAgent != userAgent {
		t.Fatalf(`Wrong UserAgent value, got "%v" instead of "%v"`, updatedFeed.UserAgent, userAgent)
	}
}

func TestUpdateFeedCookie(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	cookie := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Cookie: &cookie})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Cookie != cookie {
		t.Fatalf(`Wrong Cookie value, got "%v" instead of "%v"`, updatedFeed.Cookie, cookie)
	}

	cookie = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Cookie: &cookie})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Cookie != cookie {
		t.Fatalf(`Wrong Cookie value, got "%v" instead of "%v"`, updatedFeed.Cookie, cookie)
	}
}

func TestUpdateFeedUsername(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	username := "test"
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Username: &username})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Username != username {
		t.Fatalf(`Wrong Username value, got "%v" instead of "%v"`, updatedFeed.Username, username)
	}

	username = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Username: &username})
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
	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Password: &password})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Password != password {
		t.Fatalf(`Wrong Password value, got "%v" instead of "%v"`, updatedFeed.Password, password)
	}

	password = ""
	updatedFeed, err = client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{Password: &password})
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

	updatedFeed, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{CategoryID: &newCategory.ID})
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.Category.ID != newCategory.ID {
		t.Fatalf(`Wrong CategoryID value, got "%v" instead of "%v"`, updatedFeed.Category.ID, newCategory.ID)
	}
}

func TestUpdateFeedWithEmptyCategoryID(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	categoryID := int64(0)
	if _, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{CategoryID: &categoryID}); err == nil {
		t.Error(`Updating a feed with an empty category should not be possible`)
	}
}

func TestUpdateFeedWithInvalidCategoryID(t *testing.T) {
	client := createClient(t)
	feed, _ := createFeed(t, client)

	categoryID := int64(-1)
	if _, err := client.UpdateFeed(feed.ID, &miniflux.FeedModificationRequest{CategoryID: &categoryID}); err == nil {
		t.Error(`Updating a feed with an invalid category should not be possible`)
	}
}

func TestMarkFeedAsRead(t *testing.T) {
	client := createClient(t)

	feed, _ := createFeed(t, client)

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

	if err := client.MarkFeedAsRead(feed.ID); err != nil {
		t.Fatalf(`Failed to mark feed as read: %v`, err)
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

func TestFetchCounters(t *testing.T) {
	client := createClient(t)

	feed, _ := createFeed(t, client)

	results, err := client.FeedEntries(feed.ID, nil)
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}

	counters, err := client.FetchCounters()
	if err != nil {
		t.Fatalf(`Failed to fetch unread count: %v`, err)
	}
	unreadCounter, exists := counters.UnreadCounters[feed.ID]
	if !exists {
		unreadCounter = 0
	}

	unreadExpected := 0
	for _, entry := range results.Entries {
		if entry.Status == miniflux.EntryStatusUnread {
			unreadExpected++
		}
	}

	if unreadExpected != unreadCounter {
		t.Errorf(`Expected %d unread entries but %d instead`, unreadExpected, unreadCounter)
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

func TestGetFeedsByCategory(t *testing.T) {
	client := createClient(t)
	feed, category := createFeed(t, client)

	feeds, err := client.CategoryFeeds(category.ID)
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
