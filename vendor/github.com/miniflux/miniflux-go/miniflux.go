// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package miniflux

import (
	"fmt"
	"time"
)

// Entry statuses.
const (
	EntryStatusUnread  = "unread"
	EntryStatusRead    = "read"
	EntryStatusRemoved = "removed"
)

// User represents a user in the system.
type User struct {
	ID             int64             `json:"id"`
	Username       string            `json:"username"`
	Password       string            `json:"password,omitempty"`
	IsAdmin        bool              `json:"is_admin"`
	Theme          string            `json:"theme"`
	Language       string            `json:"language"`
	Timezone       string            `json:"timezone"`
	EntryDirection string            `json:"entry_sorting_direction"`
	LastLoginAt    *time.Time        `json:"last_login_at"`
	Extra          map[string]string `json:"extra"`
}

func (u User) String() string {
	return fmt.Sprintf("#%d - %s (admin=%v)", u.ID, u.Username, u.IsAdmin)
}

// UserModification is used to update a user.
type UserModification struct {
	Username       *string `json:"username"`
	Password       *string `json:"password"`
	IsAdmin        *bool   `json:"is_admin"`
	Theme          *string `json:"theme"`
	Language       *string `json:"language"`
	Timezone       *string `json:"timezone"`
	EntryDirection *string `json:"entry_sorting_direction"`
}

// Users represents a list of users.
type Users []User

// Category represents a category in the system.
type Category struct {
	ID     int64  `json:"id,omitempty"`
	Title  string `json:"title,omitempty"`
	UserID int64  `json:"user_id,omitempty"`
}

func (c Category) String() string {
	return fmt.Sprintf("#%d %s", c.ID, c.Title)
}

// Categories represents a list of categories.
type Categories []*Category

// Subscription represents a feed subscription.
type Subscription struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

func (s Subscription) String() string {
	return fmt.Sprintf(`Title="%s", URL="%s", Type="%s"`, s.Title, s.URL, s.Type)
}

// Subscriptions represents a list of subscriptions.
type Subscriptions []*Subscription

// Feed represents a Miniflux feed.
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
	Crawler            bool      `json:"crawler"`
	Username           string    `json:"username"`
	Password           string    `json:"password"`
	Category           *Category `json:"category,omitempty"`
	Entries            Entries   `json:"entries,omitempty"`
}

// FeedModification represents changes for a feed.
type FeedModification struct {
	FeedURL      *string `json:"feed_url"`
	SiteURL      *string `json:"site_url"`
	Title        *string `json:"title"`
	ScraperRules *string `json:"scraper_rules"`
	RewriteRules *string `json:"rewrite_rules"`
	Crawler      *bool   `json:"crawler"`
	Username     *string `json:"username"`
	Password     *string `json:"password"`
	CategoryID   *int64  `json:"category_id"`
}

// FeedIcon represents the feed icon.
type FeedIcon struct {
	ID       int64  `json:"id"`
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// Feeds represents a list of feeds.
type Feeds []*Feed

// Entry represents a subscription item in the system.
type Entry struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	FeedID     int64      `json:"feed_id"`
	Status     string     `json:"status"`
	Hash       string     `json:"hash"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	Date       time.Time  `json:"published_at"`
	Content    string     `json:"content"`
	Author     string     `json:"author"`
	Starred    bool       `json:"starred"`
	Enclosures Enclosures `json:"enclosures,omitempty"`
	Feed       *Feed      `json:"feed,omitempty"`
	Category   *Category  `json:"category,omitempty"`
}

// Entries represents a list of entries.
type Entries []*Entry

// Enclosure represents an attachment.
type Enclosure struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	EntryID  int64  `json:"entry_id"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     int    `json:"size"`
}

// Enclosures represents a list of attachments.
type Enclosures []*Enclosure

// Filter is used to filter entries.
type Filter struct {
	Status        string
	Offset        int
	Limit         int
	Order         string
	Direction     string
	Starred       bool
	Before        int64
	After         int64
	BeforeEntryID int64
	AfterEntryID  int64
}

// EntryResultSet represents the response when fetching entries.
type EntryResultSet struct {
	Total   int     `json:"total"`
	Entries Entries `json:"entries"`
}
