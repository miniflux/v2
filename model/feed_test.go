// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import (
	"testing"

	"miniflux.app/http/client"
)

func TestFeedWithResponse(t *testing.T) {
	response := &client.Response{ETag: "Some etag", LastModified: "Some date", EffectiveURL: "Some URL"}

	feed := &Feed{}
	feed.WithClientResponse(response)

	if feed.EtagHeader != "Some etag" {
		t.Fatal(`The ETag header should be set`)
	}

	if feed.LastModifiedHeader != "Some date" {
		t.Fatal(`The LastModified header should be set`)
	}

	if feed.FeedURL != "Some URL" {
		t.Fatal(`The Feed URL should be set`)
	}
}

func TestFeedCategorySetter(t *testing.T) {
	feed := &Feed{}
	feed.WithCategoryID(int64(123))

	if feed.Category == nil {
		t.Fatal(`The category field should not be null`)
	}

	if feed.Category.ID != int64(123) {
		t.Error(`The category ID must be set`)
	}
}

func TestFeedBrowsingParams(t *testing.T) {
	feed := &Feed{}
	feed.WithBrowsingParameters(true, "Custom User Agent", "Username", "Secret")

	if !feed.Crawler {
		t.Error(`The crawler must be activated`)
	}

	if feed.UserAgent != "Custom User Agent" {
		t.Error(`The user agent must be set`)
	}

	if feed.Username != "Username" {
		t.Error(`The username must be set`)
	}

	if feed.Password != "Secret" {
		t.Error(`The password must be set`)
	}
}

func TestFeedErrorCounter(t *testing.T) {
	feed := &Feed{}
	feed.WithError("Some Error")

	if feed.ParsingErrorMsg != "Some Error" {
		t.Error(`The error message must be set`)
	}

	if feed.ParsingErrorCount != 1 {
		t.Error(`The error counter must be set to 1`)
	}

	feed.ResetErrorCounter()

	if feed.ParsingErrorMsg != "" {
		t.Error(`The error message must be removed`)
	}

	if feed.ParsingErrorCount != 0 {
		t.Error(`The error counter must be set to 0`)
	}
}

func TestFeedCheckedNow(t *testing.T) {
	feed := &Feed{}
	feed.FeedURL = "https://example.org/feed"
	feed.CheckedNow()

	if feed.SiteURL != feed.FeedURL {
		t.Error(`The site URL must not be empty`)
	}

	if feed.CheckedAt.IsZero() {
		t.Error(`The checked date must be set`)
	}
}
