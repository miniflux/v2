// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestSubscriptionListEndpoint(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	fixture := servertest.NewFeedFixture(t)

	category, err := user.rest.CreateCategory("News")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := user.rest.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: fixture.FeedURL, CategoryID: category.ID})
	if err != nil {
		t.Fatal(err)
	}

	var result subscriptionListResult
	user.reader.getJSON(t, "/reader/api/0/subscription/list", nil, &result)

	if len(result.Subscriptions) != 1 {
		t.Fatalf(`Expected 1 subscription, got %+v`, result.Subscriptions)
	}
	subscription := result.Subscriptions[0]

	iconURLPrefix := config.Opts.BaseURL() + "/feed-icon/"
	if !strings.HasPrefix(subscription.IconURL, iconURLPrefix) || subscription.IconURL == iconURLPrefix {
		t.Errorf(`Expected an icon URL starting with %q, got %q`, iconURLPrefix, subscription.IconURL)
	}
	subscription.IconURL = ""

	expected := subscriptionResult{
		ID:         feedStream(feedID),
		Title:      servertest.FeedTitle,
		Categories: []tagResult{{ID: fmt.Sprintf("user/%d/label/News", user.id), Label: "News", Type: "folder"}},
		URL:        fixture.FeedURL,
		HTMLURL:    fixture.WebsiteURL,
	}
	if !reflect.DeepEqual(subscription, expected) {
		t.Errorf(`Unexpected subscription, got %+v instead of %+v`, subscription, expected)
	}
}

func TestQuickAddEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name     string
		quickAdd func(fixture *servertest.FeedFixture) string
	}{
		{"website URL", func(fixture *servertest.FeedFixture) string { return fixture.WebsiteURL }},
		{"feed URL", func(fixture *servertest.FeedFixture) string { return fixture.FeedURL }},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			user := newTestUser(t)
			fixture := servertest.NewFeedFixture(t)

			var result quickAddResult
			user.reader.postJSON(t, "/reader/api/0/subscription/quickadd", url.Values{"quickadd": {scenario.quickAdd(fixture)}}, &result)

			feeds := userFeeds(t, user)
			if len(feeds) != 1 {
				t.Fatalf(`Expected 1 feed, got %d`, len(feeds))
			}

			expected := quickAddResult{
				NumResults: 1,
				Query:      fixture.FeedURL,
				StreamID:   feedStream(feeds[0].ID),
				StreamName: servertest.FeedTitle,
			}
			if result != expected {
				t.Errorf(`Unexpected response, got %+v instead of %+v`, result, expected)
			}

			if feeds[0].FeedURL != fixture.FeedURL || feeds[0].Category.Title != "All" {
				t.Errorf(`Expected a feed for %q in the "All" category, got %q in %q`, fixture.FeedURL, feeds[0].FeedURL, feeds[0].Category.Title)
			}
		})
	}
}

func TestQuickAddFindsNoFeed(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	// The feed fixture cannot be used: the subscription finder would find its feed at a well-known URL.
	website := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(writer, `<!DOCTYPE html><html><head><title>No feed</title></head><body><p>Nothing to subscribe to.</p></body></html>`)
	}))
	t.Cleanup(website.Close)

	var result quickAddResult
	user.reader.postJSON(t, "/reader/api/0/subscription/quickadd", url.Values{"quickadd": {website.URL + "/"}}, &result)

	if result != (quickAddResult{}) {
		t.Errorf(`Expected no result, got %+v`, result)
	}
	if feeds := userFeeds(t, user); len(feeds) != 0 {
		t.Errorf(`Expected no feed, got %d`, len(feeds))
	}
}

func TestQuickAddRejectsInvalidURLs(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	for _, quickAdd := range []string{"", "/feed.xml", "example.org/feed.xml"} {
		t.Run(strconv.Quote(quickAdd), func(t *testing.T) {
			status, body := user.reader.post(t, "/reader/api/0/subscription/quickadd", url.Values{"quickadd": {quickAdd}})
			requireStatus(t, status, body, http.StatusBadRequest)
		})
	}
}

func TestQuickAddRejectsExistingFeed(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	fixture := servertest.NewFeedFixture(t)
	fixture.Subscribe(t, user.rest)

	status, body := user.reader.post(t, "/reader/api/0/subscription/quickadd", url.Values{"quickadd": {fixture.WebsiteURL}})
	requireStatus(t, status, body, http.StatusBadRequest)

	if feeds := userFeeds(t, user); len(feeds) != 1 {
		t.Errorf(`Expected 1 feed, got %d`, len(feeds))
	}
}

func TestEditSubscriptionSubscribe(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name             string
		form             url.Values
		expectedTitle    string
		expectedCategory string
	}{
		{"without title and label", url.Values{}, servertest.FeedTitle, "All"},
		{"with title and new label", url.Values{"t": {"Custom title"}, "a": {labelStream("Tech")}}, "Custom title", "Tech"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			user := newTestUser(t)
			fixture := servertest.NewFeedFixture(t)

			form := url.Values{"ac": {"subscribe"}, "s": {"feed/" + fixture.FeedURL}}
			for key, list := range scenario.form {
				form[key] = list
			}
			user.reader.postOK(t, "/reader/api/0/subscription/edit", form)

			feeds := userFeeds(t, user)
			if len(feeds) != 1 {
				t.Fatalf(`Expected 1 feed, got %d`, len(feeds))
			}
			if feeds[0].FeedURL != fixture.FeedURL {
				t.Errorf(`Expected feed URL %q, got %q`, fixture.FeedURL, feeds[0].FeedURL)
			}
			if feeds[0].Title != scenario.expectedTitle || feeds[0].Category.Title != scenario.expectedCategory {
				t.Errorf(`Expected %q in %q, got %q in %q`, scenario.expectedTitle, scenario.expectedCategory, feeds[0].Title, feeds[0].Category.Title)
			}
		})
	}
}

func TestEditSubscriptionEdit(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name             string
		form             url.Values
		expectedTitle    string
		expectedCategory string
	}{
		{"rename", url.Values{"t": {"Renamed"}}, "Renamed", "All"},
		{"move to an existing label", url.Values{"a": {labelStream("News")}}, servertest.FeedTitle, "News"},
		{"move to a new label", url.Values{"a": {labelStream("Tech")}}, servertest.FeedTitle, "Tech"},
		{"rename and move", url.Values{"t": {"Renamed"}, "a": {labelStream("News")}}, "Renamed", "News"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			user := newTestUser(t)
			feedID := servertest.NewFeedFixture(t).Subscribe(t, user.rest)
			news, err := user.rest.CreateCategory("News")
			if err != nil {
				t.Fatal(err)
			}

			form := url.Values{"ac": {"edit"}, "s": {feedStream(feedID)}}
			for key, list := range scenario.form {
				form[key] = list
			}
			user.reader.postOK(t, "/reader/api/0/subscription/edit", form)

			feed, err := user.rest.Feed(feedID)
			if err != nil {
				t.Fatal(err)
			}
			if feed.Title != scenario.expectedTitle || feed.Category.Title != scenario.expectedCategory {
				t.Errorf(`Expected %q in %q, got %q in %q`, scenario.expectedTitle, scenario.expectedCategory, feed.Title, feed.Category.Title)
			}
			if scenario.expectedCategory == "News" && feed.Category.ID != news.ID {
				t.Errorf(`Expected the existing category %d, got %d`, news.ID, feed.Category.ID)
			}
		})
	}
}

func TestEditSubscriptionUnsubscribe(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	firstFeedID := servertest.NewFeedFixture(t).Subscribe(t, user.rest)
	secondFeedID := servertest.NewFeedFixture(t).Subscribe(t, user.rest)
	keptFeedID := servertest.NewFeedFixture(t).Subscribe(t, user.rest)

	user.reader.postOK(t, "/reader/api/0/subscription/edit", url.Values{
		"ac": {"unsubscribe"},
		"s":  {feedStream(firstFeedID), feedStream(secondFeedID)},
	})

	feeds := userFeeds(t, user)
	if len(feeds) != 1 || feeds[0].ID != keptFeedID {
		t.Errorf(`Expected only feed %d to remain, got %+v`, keptFeedID, feeds)
	}
}

func TestEditSubscriptionRejectsInvalidRequests(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	fixture := servertest.NewFeedFixture(t)
	feedID := fixture.Subscribe(t, user.rest)
	unknownFeed := feedStream(feedID + 1_000_000)

	// A label whose title is the feed ID must not be taken for the feed.
	numericLabel := labelStream(strconv.FormatInt(feedID, 10))

	scenarios := []struct {
		name string
		form url.Values
	}{
		{"unknown action", url.Values{"ac": {"archive"}, "s": {feedStream(feedID)}}},
		{"missing action", url.Values{"s": {feedStream(feedID)}, "t": {"Renamed"}}},
		{"missing stream", url.Values{"ac": {"edit"}, "t": {"Renamed"}}},
		{"invalid stream", url.Values{"ac": {"edit"}, "s": {"invalid"}, "t": {"Renamed"}}},
		{"invalid label", url.Values{"ac": {"edit"}, "s": {feedStream(feedID)}, "a": {"invalid"}}},
		{"subscribe to an existing feed", url.Values{"ac": {"subscribe"}, "s": {"feed/" + fixture.FeedURL}}},
		{"subscribe to an invalid URL", url.Values{"ac": {"subscribe"}, "s": {"feed/invalid"}}},
		{"move to a state instead of a label", url.Values{"ac": {"edit"}, "s": {feedStream(feedID)}, "a": {starredStream}}},
		{"rename an unknown feed", url.Values{"ac": {"edit"}, "s": {unknownFeed}, "t": {"Renamed"}}},
		{"rename an invalid feed ID", url.Values{"ac": {"edit"}, "s": {"feed/invalid"}, "t": {"Renamed"}}},
		{"rename a label", url.Values{"ac": {"edit"}, "s": {numericLabel}, "t": {"Renamed"}}},
		{"move an unknown feed", url.Values{"ac": {"edit"}, "s": {unknownFeed}, "a": {labelStream("Tech")}}},
		{"move an invalid feed ID", url.Values{"ac": {"edit"}, "s": {"feed/invalid"}, "a": {labelStream("Tech")}}},
		{"unsubscribe from an invalid feed ID", url.Values{"ac": {"unsubscribe"}, "s": {feedStream(feedID), "feed/invalid"}}},
		{"unsubscribe from a label", url.Values{"ac": {"unsubscribe"}, "s": {numericLabel}}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			status, body := user.reader.post(t, "/reader/api/0/subscription/edit", scenario.form)
			requireStatus(t, status, body, http.StatusBadRequest)
		})
	}

	feed, err := user.rest.Feed(feedID)
	if err != nil {
		t.Fatalf(`Expected the feed to remain: %v`, err)
	}
	if feed.Title != servertest.FeedTitle || feed.Category.Title != "All" {
		t.Errorf(`Expected the feed to be unchanged, got %q in %q`, feed.Title, feed.Category.Title)
	}

	categories, err := user.rest.Categories()
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 1 {
		t.Errorf(`Expected no new category, got %+v`, categories)
	}
}

// userFeeds returns the feeds of the user, as seen by the REST API.
func userFeeds(t *testing.T, user *testUser) miniflux.Feeds {
	t.Helper()

	feeds, err := user.rest.Feeds()
	if err != nil {
		t.Fatal(err)
	}
	return feeds
}
