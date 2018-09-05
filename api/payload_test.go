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
	changes := &feedModification{FeedURL: &feedURL}
	feed := &model.Feed{FeedURL: "http://example.org/"}
	changes.Update(feed)

	if feed.FeedURL != feedURL {
		t.Fatalf(`Unexpected value, got %q instead of %q`, feed.FeedURL, feedURL)
	}
}

func TestUpdateFeedURLWithEmptyString(t *testing.T) {
	feedURL := ""
	changes := &feedModification{FeedURL: &feedURL}
	feed := &model.Feed{FeedURL: "http://example.org/"}
	changes.Update(feed)

	if feed.FeedURL == feedURL {
		t.Fatal(`The FeedURL should not be modified`)
	}
}

func TestUpdateFeedURLWhenNotSet(t *testing.T) {
	changes := &feedModification{}
	feed := &model.Feed{FeedURL: "http://example.org/"}
	changes.Update(feed)

	if feed.FeedURL != "http://example.org/" {
		t.Fatal(`The FeedURL should not be modified`)
	}
}

func TestUpdateFeedSiteURL(t *testing.T) {
	siteURL := "http://example.com/"
	changes := &feedModification{SiteURL: &siteURL}
	feed := &model.Feed{SiteURL: "http://example.org/"}
	changes.Update(feed)

	if feed.SiteURL != siteURL {
		t.Fatalf(`Unexpected value, got %q instead of %q`, feed.SiteURL, siteURL)
	}
}

func TestUpdateFeedSiteURLWithEmptyString(t *testing.T) {
	siteURL := ""
	changes := &feedModification{FeedURL: &siteURL}
	feed := &model.Feed{SiteURL: "http://example.org/"}
	changes.Update(feed)

	if feed.SiteURL == siteURL {
		t.Fatal(`The FeedURL should not be modified`)
	}
}

func TestUpdateFeedSiteURLWhenNotSet(t *testing.T) {
	changes := &feedModification{}
	feed := &model.Feed{SiteURL: "http://example.org/"}
	changes.Update(feed)

	if feed.SiteURL != "http://example.org/" {
		t.Fatal(`The SiteURL should not be modified`)
	}
}

func TestUpdateFeedTitle(t *testing.T) {
	title := "Example 2"
	changes := &feedModification{Title: &title}
	feed := &model.Feed{Title: "Example"}
	changes.Update(feed)

	if feed.Title != title {
		t.Fatalf(`Unexpected value, got %q instead of %q`, feed.Title, title)
	}
}

func TestUpdateFeedTitleWithEmptyString(t *testing.T) {
	title := ""
	changes := &feedModification{Title: &title}
	feed := &model.Feed{Title: "Example"}
	changes.Update(feed)

	if feed.Title == title {
		t.Fatal(`The Title should not be modified`)
	}
}

func TestUpdateFeedTitleWhenNotSet(t *testing.T) {
	changes := &feedModification{}
	feed := &model.Feed{Title: "Example"}
	changes.Update(feed)

	if feed.Title != "Example" {
		t.Fatal(`The Title should not be modified`)
	}
}

func TestUpdateFeedUsername(t *testing.T) {
	username := "Alice"
	changes := &feedModification{Username: &username}
	feed := &model.Feed{Username: "Bob"}
	changes.Update(feed)

	if feed.Username != username {
		t.Fatalf(`Unexpected value, got %q instead of %q`, feed.Username, username)
	}
}

func TestUpdateFeedUsernameWithEmptyString(t *testing.T) {
	username := ""
	changes := &feedModification{Username: &username}
	feed := &model.Feed{Username: "Bob"}
	changes.Update(feed)

	if feed.Username != "" {
		t.Fatal(`The Username should be empty now`)
	}
}

func TestUpdateFeedUsernameWhenNotSet(t *testing.T) {
	changes := &feedModification{}
	feed := &model.Feed{Username: "Alice"}
	changes.Update(feed)

	if feed.Username != "Alice" {
		t.Fatal(`The Username should not be modified`)
	}
}

func TestUpdateFeedCategory(t *testing.T) {
	categoryID := int64(1)
	changes := &feedModification{CategoryID: &categoryID}
	feed := &model.Feed{Category: &model.Category{ID: 42}}
	changes.Update(feed)

	if feed.Category.ID != categoryID {
		t.Fatalf(`Unexpected value, got %q instead of %q`, feed.Username, categoryID)
	}
}

func TestUpdateFeedCategoryWithZero(t *testing.T) {
	categoryID := int64(0)
	changes := &feedModification{CategoryID: &categoryID}
	feed := &model.Feed{Category: &model.Category{ID: 42}}
	changes.Update(feed)

	if feed.Category.ID != 42 {
		t.Fatal(`The CategoryID should not be modified`)
	}
}

func TestUpdateFeedCategoryWhenNotSet(t *testing.T) {
	changes := &feedModification{}
	feed := &model.Feed{Category: &model.Category{ID: 42}}
	changes.Update(feed)

	if feed.Category.ID != 42 {
		t.Fatal(`The CategoryID should not be modified`)
	}
}

func TestUpdateUserTheme(t *testing.T) {
	theme := "Example 2"
	changes := &userModification{Theme: &theme}
	user := &model.User{Theme: "Example"}
	changes.Update(user)

	if user.Theme != theme {
		t.Fatalf(`Unexpected value, got %q instead of %q`, user.Theme, theme)
	}
}

func TestUserThemeWhenNotSet(t *testing.T) {
	changes := &userModification{}
	user := &model.User{Theme: "Example"}
	changes.Update(user)

	if user.Theme != "Example" {
		t.Fatal(`The user Theme should not be modified`)
	}
}
