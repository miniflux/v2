// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
	"reflect"
	"time"
)

// Feed represents a feed in the database.
type Feed struct {
	ID                 int64     `json:"id"`
	UserID             int64     `json:"user_id"`
	FeedURL            string    `json:"feed_url"`
	SiteURL            string    `json:"site_url"`
	Title              string    `json:"title"`
	CheckedAt          time.Time `json:"checked_at,omitempty"`
	EtagHeader         string    `json:"etag_header,omitempty"`
	LastModifiedHeader string    `json:"last_modified_header,omitempty"`
	ParsingErrorMsg    string    `json:"parsing_error_message,omitempty"`
	ParsingErrorCount  int       `json:"parsing_error_count,omitempty"`
	ScraperRules       string    `json:"scraper_rules"`
	RewriteRules       string    `json:"rewrite_rules"`
	Category           *Category `json:"category,omitempty"`
	Entries            Entries   `json:"entries,omitempty"`
	Icon               *FeedIcon `json:"icon,omitempty"`
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

// Merge combine src to the current struct
func (f *Feed) Merge(src *Feed) {
	src.ID = f.ID
	src.UserID = f.UserID

	new := reflect.ValueOf(src).Elem()
	for i := 0; i < new.NumField(); i++ {
		field := new.Field(i)

		switch field.Interface().(type) {
		case int64:
			value := field.Int()
			if value != 0 {
				reflect.ValueOf(f).Elem().Field(i).SetInt(value)
			}
		case string:
			value := field.String()
			if value != "" {
				reflect.ValueOf(f).Elem().Field(i).SetString(value)
			}
		}
	}
}

// Feeds is a list of feed
type Feeds []*Feed
