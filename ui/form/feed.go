// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form

import (
	"net/http"
	"strconv"

	"github.com/miniflux/miniflux/errors"
	"github.com/miniflux/miniflux/model"
)

// FeedForm represents a feed form in the UI
type FeedForm struct {
	FeedURL      string
	SiteURL      string
	Title        string
	ScraperRules string
	RewriteRules string
	Crawler      bool
	CategoryID   int64
}

// ValidateModification validates FeedForm fields
func (f FeedForm) ValidateModification() error {
	if f.FeedURL == "" || f.SiteURL == "" || f.Title == "" || f.CategoryID == 0 {
		return errors.NewLocalizedError("All fields are mandatory.")
	}
	return nil
}

// Merge updates the fields of the given feed.
func (f FeedForm) Merge(feed *model.Feed) *model.Feed {
	feed.Category.ID = f.CategoryID
	feed.Title = f.Title
	feed.SiteURL = f.SiteURL
	feed.FeedURL = f.FeedURL
	feed.ScraperRules = f.ScraperRules
	feed.RewriteRules = f.RewriteRules
	feed.Crawler = f.Crawler
	feed.ParsingErrorCount = 0
	feed.ParsingErrorMsg = ""
	return feed
}

// NewFeedForm parses the HTTP request and returns a FeedForm
func NewFeedForm(r *http.Request) *FeedForm {
	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil {
		categoryID = 0
	}

	return &FeedForm{
		FeedURL:      r.FormValue("feed_url"),
		SiteURL:      r.FormValue("site_url"),
		Title:        r.FormValue("title"),
		ScraperRules: r.FormValue("scraper_rules"),
		RewriteRules: r.FormValue("rewrite_rules"),
		Crawler:      r.FormValue("crawler") == "1",
		CategoryID:   int64(categoryID),
	}
}
