// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package client // import "miniflux.app/client"

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
	ID                int64      `json:"id"`
	Username          string     `json:"username"`
	Password          string     `json:"password,omitempty"`
	IsAdmin           bool       `json:"is_admin"`
	Theme             string     `json:"theme"`
	Language          string     `json:"language"`
	Timezone          string     `json:"timezone"`
	EntryDirection    string     `json:"entry_sorting_direction"`
	Stylesheet        string     `json:"stylesheet"`
	GoogleID          string     `json:"google_id"`
	OpenIDConnectID   string     `json:"openid_connect_id"`
	EntriesPerPage    int        `json:"entries_per_page"`
	KeyboardShortcuts bool       `json:"keyboard_shortcuts"`
	ShowReadingTime   bool       `json:"show_reading_time"`
	EntrySwipe        bool       `json:"entry_swipe"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	DisplayMode       string     `json:"display_mode"`
}

func (u User) String() string {
	return fmt.Sprintf("#%d - %s (admin=%v)", u.ID, u.Username, u.IsAdmin)
}

// UserCreationRequest represents the request to create a user.
type UserCreationRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	IsAdmin         bool   `json:"is_admin"`
	GoogleID        string `json:"google_id"`
	OpenIDConnectID string `json:"openid_connect_id"`
}

// UserModificationRequest represents the request to update a user.
type UserModificationRequest struct {
	Username          *string `json:"username"`
	Password          *string `json:"password"`
	IsAdmin           *bool   `json:"is_admin"`
	Theme             *string `json:"theme"`
	Language          *string `json:"language"`
	Timezone          *string `json:"timezone"`
	EntryDirection    *string `json:"entry_sorting_direction"`
	Stylesheet        *string `json:"stylesheet"`
	GoogleID          *string `json:"google_id"`
	OpenIDConnectID   *string `json:"openid_connect_id"`
	EntriesPerPage    *int    `json:"entries_per_page"`
	KeyboardShortcuts *bool   `json:"keyboard_shortcuts"`
	ShowReadingTime   *bool   `json:"show_reading_time"`
	EntrySwipe        *bool   `json:"entry_swipe"`
	DisplayMode       *string `json:"display_mode"`
}

// Users represents a list of users.
type Users []User

// Category represents a feed category.
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
	ID                          int64     `json:"id"`
	UserID                      int64     `json:"user_id"`
	FeedURL                     string    `json:"feed_url"`
	SiteURL                     string    `json:"site_url"`
	Title                       string    `json:"title"`
	CheckedAt                   time.Time `json:"checked_at,omitempty"`
	EtagHeader                  string    `json:"etag_header,omitempty"`
	LastModifiedHeader          string    `json:"last_modified_header,omitempty"`
	ParsingErrorMsg             string    `json:"parsing_error_message,omitempty"`
	ParsingErrorCount           int       `json:"parsing_error_count,omitempty"`
	Disabled                    bool      `json:"disabled"`
	IgnoreHTTPCache             bool      `json:"ignore_http_cache"`
	AllowSelfSignedCertificates bool      `json:"allow_self_signed_certificates"`
	FetchViaProxy               bool      `json:"fetch_via_proxy"`
	ScraperRules                string    `json:"scraper_rules"`
	RewriteRules                string    `json:"rewrite_rules"`
	BlocklistRules              string    `json:"blocklist_rules"`
	KeeplistRules               string    `json:"keeplist_rules"`
	Crawler                     bool      `json:"crawler"`
	UserAgent                   string    `json:"user_agent"`
	Cookie                      string    `json:"cookie"`
	Username                    string    `json:"username"`
	Password                    string    `json:"password"`
	Category                    *Category `json:"category,omitempty"`
}

// FeedCreationRequest represents the request to create a feed.
type FeedCreationRequest struct {
	FeedURL                     string `json:"feed_url"`
	CategoryID                  int64  `json:"category_id"`
	UserAgent                   string `json:"user_agent"`
	Cookie                      string `json:"cookie"`
	Username                    string `json:"username"`
	Password                    string `json:"password"`
	Crawler                     bool   `json:"crawler"`
	Disabled                    bool   `json:"disabled"`
	IgnoreHTTPCache             bool   `json:"ignore_http_cache"`
	AllowSelfSignedCertificates bool   `json:"allow_self_signed_certificates"`
	FetchViaProxy               bool   `json:"fetch_via_proxy"`
	ScraperRules                string `json:"scraper_rules"`
	RewriteRules                string `json:"rewrite_rules"`
	BlocklistRules              string `json:"blocklist_rules"`
	KeeplistRules               string `json:"keeplist_rules"`
}

// FeedModificationRequest represents the request to update a feed.
type FeedModificationRequest struct {
	FeedURL                     *string `json:"feed_url"`
	SiteURL                     *string `json:"site_url"`
	Title                       *string `json:"title"`
	ScraperRules                *string `json:"scraper_rules"`
	RewriteRules                *string `json:"rewrite_rules"`
	BlocklistRules              *string `json:"blocklist_rules"`
	KeeplistRules               *string `json:"keeplist_rules"`
	Crawler                     *bool   `json:"crawler"`
	UserAgent                   *string `json:"user_agent"`
	Cookie                      *string `json:"cookie"`
	Username                    *string `json:"username"`
	Password                    *string `json:"password"`
	CategoryID                  *int64  `json:"category_id"`
	Disabled                    *bool   `json:"disabled"`
	IgnoreHTTPCache             *bool   `json:"ignore_http_cache"`
	AllowSelfSignedCertificates *bool   `json:"allow_self_signed_certificates"`
	FetchViaProxy               *bool   `json:"fetch_via_proxy"`
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
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	FeedID      int64      `json:"feed_id"`
	Status      string     `json:"status"`
	Hash        string     `json:"hash"`
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	CommentsURL string     `json:"comments_url"`
	Date        time.Time  `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	ChangedAt   time.Time  `json:"changed_at"`
	Content     string     `json:"content"`
	Author      string     `json:"author"`
	ShareCode   string     `json:"share_code"`
	Starred     bool       `json:"starred"`
	ReadingTime int        `json:"reading_time"`
	Enclosures  Enclosures `json:"enclosures,omitempty"`
	Feed        *Feed      `json:"feed,omitempty"`
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
	Search        string
	CategoryID    int64
	FeedID        int64
	Statuses      []string
}

// EntryResultSet represents the response when fetching entries.
type EntryResultSet struct {
	Total   int     `json:"total"`
	Entries Entries `json:"entries"`
}
