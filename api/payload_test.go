// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"testing"

	"miniflux.app/model"
)

func TestUpdateFeedURL(t *testing.T) {
	feedURL := "http://example.com/"
	changes := &feedModificationRequest{FeedURL: &feedURL}
	feed := &model.Feed{FeedURL: "http://example.org/"}
	changes.Update(feed)

	if feed.FeedURL != feedURL {
		t.Errorf(`Unexpected value, got %q instead of %q`, feed.FeedURL, feedURL)
	}
}

func TestUpdateFeedURLWithEmptyString(t *testing.T) {
	feedURL := ""
	changes := &feedModificationRequest{FeedURL: &feedURL}
	feed := &model.Feed{FeedURL: "http://example.org/"}
	changes.Update(feed)

	if feed.FeedURL == feedURL {
		t.Error(`The FeedURL should not be modified`)
	}
}

func TestUpdateFeedURLWhenNotSet(t *testing.T) {
	changes := &feedModificationRequest{}
	feed := &model.Feed{FeedURL: "http://example.org/"}
	changes.Update(feed)

	if feed.FeedURL != "http://example.org/" {
		t.Error(`The FeedURL should not be modified`)
	}
}

func TestUpdateFeedSiteURL(t *testing.T) {
	siteURL := "http://example.com/"
	changes := &feedModificationRequest{SiteURL: &siteURL}
	feed := &model.Feed{SiteURL: "http://example.org/"}
	changes.Update(feed)

	if feed.SiteURL != siteURL {
		t.Errorf(`Unexpected value, got %q instead of %q`, feed.SiteURL, siteURL)
	}
}

func TestUpdateFeedSiteURLWithEmptyString(t *testing.T) {
	siteURL := ""
	changes := &feedModificationRequest{FeedURL: &siteURL}
	feed := &model.Feed{SiteURL: "http://example.org/"}
	changes.Update(feed)

	if feed.SiteURL == siteURL {
		t.Error(`The FeedURL should not be modified`)
	}
}

func TestUpdateFeedSiteURLWhenNotSet(t *testing.T) {
	changes := &feedModificationRequest{}
	feed := &model.Feed{SiteURL: "http://example.org/"}
	changes.Update(feed)

	if feed.SiteURL != "http://example.org/" {
		t.Error(`The SiteURL should not be modified`)
	}
}

func TestUpdateFeedTitle(t *testing.T) {
	title := "Example 2"
	changes := &feedModificationRequest{Title: &title}
	feed := &model.Feed{Title: "Example"}
	changes.Update(feed)

	if feed.Title != title {
		t.Errorf(`Unexpected value, got %q instead of %q`, feed.Title, title)
	}
}

func TestUpdateFeedTitleWithEmptyString(t *testing.T) {
	title := ""
	changes := &feedModificationRequest{Title: &title}
	feed := &model.Feed{Title: "Example"}
	changes.Update(feed)

	if feed.Title == title {
		t.Error(`The Title should not be modified`)
	}
}

func TestUpdateFeedTitleWhenNotSet(t *testing.T) {
	changes := &feedModificationRequest{}
	feed := &model.Feed{Title: "Example"}
	changes.Update(feed)

	if feed.Title != "Example" {
		t.Error(`The Title should not be modified`)
	}
}

func TestUpdateFeedUsername(t *testing.T) {
	username := "Alice"
	changes := &feedModificationRequest{Username: &username}
	feed := &model.Feed{Username: "Bob"}
	changes.Update(feed)

	if feed.Username != username {
		t.Errorf(`Unexpected value, got %q instead of %q`, feed.Username, username)
	}
}

func TestUpdateFeedUsernameWithEmptyString(t *testing.T) {
	username := ""
	changes := &feedModificationRequest{Username: &username}
	feed := &model.Feed{Username: "Bob"}
	changes.Update(feed)

	if feed.Username != "" {
		t.Error(`The Username should be empty now`)
	}
}

func TestUpdateFeedUsernameWhenNotSet(t *testing.T) {
	changes := &feedModificationRequest{}
	feed := &model.Feed{Username: "Alice"}
	changes.Update(feed)

	if feed.Username != "Alice" {
		t.Error(`The Username should not be modified`)
	}
}

func TestUpdateFeedDisabled(t *testing.T) {
	valueTrue := true
	valueFalse := false
	scenarios := []struct {
		changes  *feedModificationRequest
		feed     *model.Feed
		expected bool
	}{
		{&feedModificationRequest{}, &model.Feed{Disabled: true}, true},
		{&feedModificationRequest{Disabled: &valueTrue}, &model.Feed{Disabled: true}, true},
		{&feedModificationRequest{Disabled: &valueFalse}, &model.Feed{Disabled: true}, false},
		{&feedModificationRequest{}, &model.Feed{Disabled: false}, false},
		{&feedModificationRequest{Disabled: &valueTrue}, &model.Feed{Disabled: false}, true},
		{&feedModificationRequest{Disabled: &valueFalse}, &model.Feed{Disabled: false}, false},
	}

	for _, scenario := range scenarios {
		scenario.changes.Update(scenario.feed)
		if scenario.feed.Disabled != scenario.expected {
			t.Errorf(`Unexpected result, got %v, want: %v`,
				scenario.feed.Disabled,
				scenario.expected,
			)
		}
	}
}

func TestUpdateFeedCategory(t *testing.T) {
	categoryID := int64(1)
	changes := &feedModificationRequest{CategoryID: &categoryID}
	feed := &model.Feed{Category: &model.Category{ID: 42}}
	changes.Update(feed)

	if feed.Category.ID != categoryID {
		t.Errorf(`Unexpected value, got %q instead of %q`, feed.Username, categoryID)
	}
}

func TestUpdateFeedCategoryWithZero(t *testing.T) {
	categoryID := int64(0)
	changes := &feedModificationRequest{CategoryID: &categoryID}
	feed := &model.Feed{Category: &model.Category{ID: 42}}
	changes.Update(feed)

	if feed.Category.ID != 42 {
		t.Error(`The CategoryID should not be modified`)
	}
}

func TestUpdateFeedCategoryWhenNotSet(t *testing.T) {
	changes := &feedModificationRequest{}
	feed := &model.Feed{Category: &model.Category{ID: 42}}
	changes.Update(feed)

	if feed.Category.ID != 42 {
		t.Error(`The CategoryID should not be modified`)
	}
}

func TestUpdateFeedToIgnoreCache(t *testing.T) {
	value := true
	changes := &feedModificationRequest{IgnoreHTTPCache: &value}
	feed := &model.Feed{IgnoreHTTPCache: false}
	changes.Update(feed)

	if feed.IgnoreHTTPCache != value {
		t.Errorf(`The field IgnoreHTTPCache should be %v`, value)
	}
}

func TestUpdateFeedToFetchViaProxy(t *testing.T) {
	value := true
	changes := &feedModificationRequest{FetchViaProxy: &value}
	feed := &model.Feed{FetchViaProxy: false}
	changes.Update(feed)

	if feed.FetchViaProxy != value {
		t.Errorf(`The field FetchViaProxy should be %v`, value)
	}
}
