// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
	"time"
)

// Feed represents a feed in the database.
type Feed struct {
	ID                 int64     `json:"id"`
	UserID             int64     `json:"user_id"`
	FeedURL            string    `json:"feed_url"`
	SiteURL            string    `json:"site_url"`
	Title              string    `json:"title"`
	CheckedAt          time.Time `json:"checked_at"`
	EtagHeader         string    `json:"etag_header"`
	LastModifiedHeader string    `json:"last_modified_header"`
	ParsingErrorMsg    string    `json:"parsing_error_message"`
	ParsingErrorCount  int       `json:"parsing_error_count"`
	ScraperRules       string    `json:"scraper_rules"`
	RewriteRules       string    `json:"rewrite_rules"`
	Crawler            bool      `json:"crawler"`
	Username           string    `json:"username"`
	Password           string    `json:"password"`
	Category           *Category `json:"category,omitempty"`
	Entries            Entries   `json:"entries,omitempty"`
	Icon               *FeedIcon `json:"icon"`
}

func (f *Feed) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, FeedURL=%s, SiteURL=%s, Title=%s, Category={%s}",
		f.ID,
		f.UserID,
		f.FeedURL,
		f.SiteURL,
		f.Title,
		f.Category,
	)
}

// Feeds is a list of feed
type Feeds []*Feed
