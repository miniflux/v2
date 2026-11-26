// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestCreateFeedEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)
	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    fixture.feedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Errorf(`Invalid feedID, got "%v"`, feedID)
	}
}

func TestCannotCreateDuplicatedFeed(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	if feedID == 0 {
		t.Fatalf(`Invalid feedID, got "%v"`, feedID)
	}

	_, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: fixture.feedURL,
	})
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Duplicated feeds should not be allowed, got %v`, err)
	}
}

func TestFeedCreationWithoutCategoryReturnsBadRequest(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)
	categories, err := regularUserClient.Categories()
	if err != nil {
		t.Fatal(err)
	}

	for _, category := range categories {
		if err := regularUserClient.DeleteCategory(category.ID); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("REST API", func(t *testing.T) {
		_, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: fixture.feedURL})
		if !errors.Is(err, miniflux.ErrBadRequest) {
			t.Fatalf(`Expected a bad request error, got %v`, err)
		}
	})

	t.Run("OPML import", func(t *testing.T) {
		data := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
    <body>
        <outline title="Test" text="Test" xmlUrl="` + fixture.feedURL + `"></outline>
    </body>
</opml>`

		err := regularUserClient.Import(io.NopCloser(bytes.NewReader([]byte(data))))
		if !errors.Is(err, miniflux.ErrBadRequest) {
			t.Fatalf(`Expected a bad request error, got %v`, err)
		}
	})
}

func TestCreateFeedWithInexistingCategory(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	_, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    fixture.feedURL,
		CategoryID: 123456789,
	})

	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating a feed with an inexisting category should raise an error, got %v`, err)
	}
}

func TestCreateFeedWithEmptyFeedURL(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := testConfig.newAdminClient()
	_, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: "",
	})
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating a feed with an empty feed URL should raise an error, got %v`, err)
	}
}

func TestCreateFeedWithInvalidFeedURL(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := testConfig.newAdminClient()
	_, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: "invalid_feed_url",
	})
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating a feed with an invalid feed URL should raise an error, got %v`, err)
	}
}

func TestCreateDisabledFeed(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:  fixture.feedURL,
		Disabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	feed, err := regularUserClient.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if !feed.Disabled {
		t.Fatalf(`The feed should be disabled`)
	}
}

func TestCreateFeedWithDisabledHTTPCache(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:         fixture.feedURL,
		IgnoreHTTPCache: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	feed, err := regularUserClient.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if !feed.IgnoreHTTPCache {
		t.Fatalf(`The feed should ignore the HTTP cache`)
	}
}

func TestCreateFeedWithScraperRule(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:      fixture.feedURL,
		ScraperRules: "article",
	})
	if err != nil {
		t.Fatal(err)
	}

	feed, err := regularUserClient.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if feed.ScraperRules != "article" {
		t.Fatalf(`The feed should have the scraper rules set to "article"`)
	}
}

func TestUpdateFeedEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		FeedURL: new("https://example.org/feed.xml"),
	}

	updatedFeed, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.FeedURL != "https://example.org/feed.xml" {
		t.Fatalf(`Invalid feed URL, got "%v"`, updatedFeed.FeedURL)
	}
}

func TestCannotHaveDuplicateFeedWhenUpdatingFeed(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)
	otherFixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	fixture.subscribe(t, regularUserClient)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: otherFixture.feedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		FeedURL: new(fixture.feedURL),
	}

	if _, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Duplicated feeds should not be allowed, got %v`, err)
	}
}

func TestUpdateFeedWithInvalidCategory(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		CategoryID: new(int64(123456789)),
	}

	if _, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Updating a feed with an inexisting category should raise an error, got %v`, err)
	}
}

func TestCreateFeedWithInvalidRewriteRules(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	// Missing the second backslash in the search term makes this an invalid Go
	// string literal, which the rewrite rule parser must reject.
	_, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:      fixture.feedURL,
		RewriteRules: `replace("\d+"|"x")`,
	})
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating a feed with a malformed rewrite rule should raise an error, got %v`, err)
	}
}

func TestUpdateFeedWithInvalidRewriteRules(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		RewriteRules: new(`replace("\d+"|"x")`),
	}

	if _, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Updating a feed with a malformed rewrite rule should raise an error, got %v`, err)
	}
}

func TestCreateFeedWithInvalidURLRewriteRules(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	// "[" is an unterminated character class and therefore an invalid regexp.
	_, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:         fixture.feedURL,
		UrlRewriteRules: "[",
	})
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating a feed with an invalid URL rewrite rule should raise an error, got %v`, err)
	}
}

func TestUpdateFeedWithInvalidURLRewriteRules(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		UrlRewriteRules: new("["),
	}

	if _, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Updating a feed with an invalid URL rewrite rule should raise an error, got %v`, err)
	}
}

func TestCreateFeedWithInvalidRewriteRuleSelector(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	// cascadia does not support a relative selector inside :has().
	_, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:      fixture.feedURL,
		RewriteRules: `remove("div:has(> p)")`,
	})
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating a feed with an invalid rewrite rule selector should raise an error, got %v`, err)
	}
}

func TestUpdateFeedWithInvalidRewriteRuleSelector(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		RewriteRules: new(`remove("div:has(> p)")`),
	}

	if _, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Updating a feed with an invalid rewrite rule selector should raise an error, got %v`, err)
	}
}

func TestCreateFeedWithInvalidScraperRules(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	// cascadia does not support a relative selector inside :has().
	_, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:      fixture.feedURL,
		ScraperRules: `div:has(> p)`,
	})
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating a feed with invalid scraper rules should raise an error, got %v`, err)
	}
}

func TestUpdateFeedWithInvalidScraperRules(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		ScraperRules: new(`div:has(> p)`),
	}

	if _, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Updating a feed with invalid scraper rules should raise an error, got %v`, err)
	}
}

func TestMarkFeedAsReadEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	if err := regularUserClient.MarkFeedAsRead(feedID); err != nil {
		t.Fatal(err)
	}

	results := listFeedEntries(t, regularUserClient, feedID, nil)

	for _, entry := range results.Entries {
		if entry.Status != miniflux.EntryStatusRead {
			t.Errorf(`Status for entry %d was %q instead of %q`, entry.ID, entry.Status, miniflux.EntryStatusRead)
		}
	}
}

func TestFetchCountersEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	counters, err := regularUserClient.FetchCounters()
	if err != nil {
		t.Fatal(err)
	}

	if value, ok := counters.ReadCounters[feedID]; ok && value != 0 {
		t.Errorf(`Invalid read counter, got %d`, value)
	}

	if value, ok := counters.UnreadCounters[feedID]; !ok || value == 0 {
		t.Errorf(`Invalid unread counter, got %d`, value)
	}
}

func TestDeleteFeedEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	if err := regularUserClient.DeleteFeed(feedID); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshAllFeedsEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	if err := regularUserClient.RefreshAllFeeds(); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshFeedEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	if err := regularUserClient.RefreshFeed(feedID); err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.ForceRefreshFeed(feedID); err != nil {
		t.Fatal(err)
	}
}

func TestGetFeedEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	feed, err := regularUserClient.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if feed.ID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, feed.ID)
	}

	if feed.FeedURL != fixture.feedURL {
		t.Fatalf(`Invalid feed URL, got %q`, feed.FeedURL)
	}

	if feed.SiteURL != fixture.websiteURL {
		t.Fatalf(`Invalid site URL, got %q`, feed.SiteURL)
	}

	if feed.Title != fixtureFeedTitle {
		t.Fatalf(`Invalid title, got %q`, feed.Title)
	}
}

func TestGetFeedsEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID := fixture.subscribe(t, regularUserClient)

	feeds, err := regularUserClient.Feeds()
	if err != nil {
		t.Fatal(err)
	}

	if len(feeds) != 1 {
		t.Fatalf(`Invalid number of feeds, got %d`, len(feeds))
	}

	if feeds[0].ID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, feeds[0].ID)
	}

	if feeds[0].FeedURL != fixture.feedURL {
		t.Fatalf(`Invalid feed URL, got %q`, feeds[0].FeedURL)
	}
}

func TestGetCategoryFeedsEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    fixture.feedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	feeds, err := regularUserClient.CategoryFeeds(category.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(feeds) != 1 {
		t.Fatalf(`Invalid number of feeds, got %d`, len(feeds))
	}

	if feeds[0].ID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, feeds[0].ID)
	}

	if feeds[0].FeedURL != fixture.feedURL {
		t.Fatalf(`Invalid feed URL, got %q`, feeds[0].FeedURL)
	}
}
