// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client // import "miniflux.app/v2/client"

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
	ID                        int64      `json:"id"`
	Username                  string     `json:"username"`
	Password                  string     `json:"password,omitempty"`
	IsAdmin                   bool       `json:"is_admin"`
	Theme                     string     `json:"theme"`
	Language                  string     `json:"language"`
	Timezone                  string     `json:"timezone"`
	EntryDirection            string     `json:"entry_sorting_direction"`
	EntryOrder                string     `json:"entry_sorting_order"`
	Stylesheet                string     `json:"stylesheet"`
	CustomJS                  string     `json:"custom_js"`
	GoogleID                  string     `json:"google_id"`
	OpenIDConnectID           string     `json:"openid_connect_id"`
	EntriesPerPage            int        `json:"entries_per_page"`
	KeyboardShortcuts         bool       `json:"keyboard_shortcuts"`
	ShowReadingTime           bool       `json:"show_reading_time"`
	EntrySwipe                bool       `json:"entry_swipe"`
	GestureNav                string     `json:"gesture_nav"`
	LastLoginAt               *time.Time `json:"last_login_at"`
	DisplayMode               string     `json:"display_mode"`
	DefaultReadingSpeed       int        `json:"default_reading_speed"`
	CJKReadingSpeed           int        `json:"cjk_reading_speed"`
	DefaultHomePage           string     `json:"default_home_page"`
	CategoriesSortingOrder    string     `json:"categories_sorting_order"`
	MarkReadOnView            bool       `json:"mark_read_on_view"`
	MediaPlaybackRate         float64    `json:"media_playback_rate"`
	BlockFilterEntryRules     string     `json:"block_filter_entry_rules"`
	KeepFilterEntryRules      string     `json:"keep_filter_entry_rules"`
	ExternalFontHosts         string     `json:"external_font_hosts"`
	AlwaysOpenExternalLinks   bool       `json:"always_open_external_links"`
	OpenExternalLinksInNewTab bool       `json:"open_external_links_in_new_tab"`
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
	Username                  *string  `json:"username"`
	Password                  *string  `json:"password"`
	IsAdmin                   *bool    `json:"is_admin"`
	Theme                     *string  `json:"theme"`
	Language                  *string  `json:"language"`
	Timezone                  *string  `json:"timezone"`
	EntryDirection            *string  `json:"entry_sorting_direction"`
	EntryOrder                *string  `json:"entry_sorting_order"`
	Stylesheet                *string  `json:"stylesheet"`
	CustomJS                  *string  `json:"custom_js"`
	GoogleID                  *string  `json:"google_id"`
	OpenIDConnectID           *string  `json:"openid_connect_id"`
	EntriesPerPage            *int     `json:"entries_per_page"`
	KeyboardShortcuts         *bool    `json:"keyboard_shortcuts"`
	ShowReadingTime           *bool    `json:"show_reading_time"`
	EntrySwipe                *bool    `json:"entry_swipe"`
	GestureNav                *string  `json:"gesture_nav"`
	DisplayMode               *string  `json:"display_mode"`
	DefaultReadingSpeed       *int     `json:"default_reading_speed"`
	CJKReadingSpeed           *int     `json:"cjk_reading_speed"`
	DefaultHomePage           *string  `json:"default_home_page"`
	CategoriesSortingOrder    *string  `json:"categories_sorting_order"`
	MarkReadOnView            *bool    `json:"mark_read_on_view"`
	MediaPlaybackRate         *float64 `json:"media_playback_rate"`
	BlockFilterEntryRules     *string  `json:"block_filter_entry_rules"`
	KeepFilterEntryRules      *string  `json:"keep_filter_entry_rules"`
	ExternalFontHosts         *string  `json:"external_font_hosts"`
	AlwaysOpenExternalLinks   *bool    `json:"always_open_external_links"`
	OpenExternalLinksInNewTab *bool    `json:"open_external_links_in_new_tab"`
}

// Users represents a list of users.
type Users []User

// Category represents a feed category.
type Category struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	UserID       int64  `json:"user_id,omitempty"`
	HideGlobally bool   `json:"hide_globally,omitempty"`
	FeedCount    *int   `json:"feed_count,omitempty"`
	TotalUnread  *int   `json:"total_unread,omitempty"`
}

func (c Category) String() string {
	return fmt.Sprintf("#%d %s", c.ID, c.Title)
}

// Categories represents a list of categories.
type Categories []*Category

// CategoryCreationRequest represents the request to create a category.
type CategoryCreationRequest struct {
	Title        string `json:"title"`
	HideGlobally bool   `json:"hide_globally"`
}

// CategoryModificationRequest represents the request to update a category.
type CategoryModificationRequest struct {
	Title        *string `json:"title"`
	HideGlobally *bool   `json:"hide_globally"`
}

// Subscription represents a feed subscription.
type Subscription struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

func (s Subscription) String() string {
	return fmt.Sprintf(`Title=%q, URL=%q, Type=%q`, s.Title, s.URL, s.Type)
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
	UrlRewriteRules             string    `json:"urlrewrite_rules"`
	BlocklistRules              string    `json:"blocklist_rules"`
	KeeplistRules               string    `json:"keeplist_rules"`
	BlockFilterEntryRules       string    `json:"block_filter_entry_rules"`
	KeepFilterEntryRules        string    `json:"keep_filter_entry_rules"`
	Crawler                     bool      `json:"crawler"`
	UserAgent                   string    `json:"user_agent"`
	Cookie                      string    `json:"cookie"`
	Username                    string    `json:"username"`
	Password                    string    `json:"password"`
	Category                    *Category `json:"category,omitempty"`
	HideGlobally                bool      `json:"hide_globally"`
	DisableHTTP2                bool      `json:"disable_http2"`
	ProxyURL                    string    `json:"proxy_url"`
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
	UrlRewriteRules             string `json:"urlrewrite_rules"`
	BlocklistRules              string `json:"blocklist_rules"`
	KeeplistRules               string `json:"keeplist_rules"`
	BlockFilterEntryRules       string `json:"block_filter_entry_rules"`
	KeepFilterEntryRules        string `json:"keep_filter_entry_rules"`
	HideGlobally                bool   `json:"hide_globally"`
	DisableHTTP2                bool   `json:"disable_http2"`
	ProxyURL                    string `json:"proxy_url"`
}

// FeedModificationRequest represents the request to update a feed.
type FeedModificationRequest struct {
	FeedURL                     *string `json:"feed_url"`
	SiteURL                     *string `json:"site_url"`
	Title                       *string `json:"title"`
	ScraperRules                *string `json:"scraper_rules"`
	RewriteRules                *string `json:"rewrite_rules"`
	UrlRewriteRules             *string `json:"urlrewrite_rules"`
	BlocklistRules              *string `json:"blocklist_rules"`
	KeeplistRules               *string `json:"keeplist_rules"`
	BlockFilterEntryRules       *string `json:"block_filter_entry_rules"`
	KeepFilterEntryRules        *string `json:"keep_filter_entry_rules"`
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
	HideGlobally                *bool   `json:"hide_globally"`
	DisableHTTP2                *bool   `json:"disable_http2"`
	ProxyURL                    *string `json:"proxy_url"`
}

// FeedIcon represents the feed icon.
type FeedIcon struct {
	ID       int64  `json:"id"`
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type FeedCounters struct {
	ReadCounters   map[int64]int `json:"reads"`
	UnreadCounters map[int64]int `json:"unreads"`
}

// Feeds represents a list of feeds.
type Feeds []*Feed

// Entry represents a subscription item in the system.
type Entry struct {
	ID          int64      `json:"id"`
	Date        time.Time  `json:"published_at"`
	ChangedAt   time.Time  `json:"changed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Feed        *Feed      `json:"feed,omitempty"`
	Hash        string     `json:"hash"`
	URL         string     `json:"url"`
	CommentsURL string     `json:"comments_url"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	Content     string     `json:"content"`
	Author      string     `json:"author"`
	ShareCode   string     `json:"share_code"`
	Enclosures  Enclosures `json:"enclosures,omitempty"`
	Tags        []string   `json:"tags"`
	ReadingTime int        `json:"reading_time"`
	UserID      int64      `json:"user_id"`
	FeedID      int64      `json:"feed_id"`
	Starred     bool       `json:"starred"`
}

// EntryModificationRequest represents a request to modify an entry.
type EntryModificationRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

// Entries represents a list of entries.
type Entries []*Entry

// Enclosure represents an attachment.
type Enclosure struct {
	ID               int64  `json:"id"`
	UserID           int64  `json:"user_id"`
	EntryID          int64  `json:"entry_id"`
	URL              string `json:"url"`
	MimeType         string `json:"mime_type"`
	Size             int    `json:"size"`
	MediaProgression int64  `json:"media_progression"`
}

type EnclosureUpdateRequest struct {
	MediaProgression int64 `json:"media_progression"`
}

// Enclosures represents a list of attachments.
type Enclosures []*Enclosure

const (
	FilterNotStarred  = "0"
	FilterOnlyStarred = "1"
)

// Filter is used to filter entries.
type Filter struct {
	Status          string
	Offset          int
	Limit           int
	Order           string
	Direction       string
	Starred         string
	Before          int64
	After           int64
	PublishedBefore int64
	PublishedAfter  int64
	ChangedBefore   int64
	ChangedAfter    int64
	BeforeEntryID   int64
	AfterEntryID    int64
	Search          string
	CategoryID      int64
	FeedID          int64
	Statuses        []string
	GloballyVisible bool
}

// EntryResultSet represents the response when fetching entries.
type EntryResultSet struct {
	Total   int     `json:"total"`
	Entries Entries `json:"entries"`
}

// VersionResponse represents the version and the build information of the Miniflux instance.
type VersionResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Compiler  string `json:"compiler"`
	Arch      string `json:"arch"`
	OS        string `json:"os"`
}

// APIKey represents an application API key.
type APIKey struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Token       string     `json:"token"`
	Description string     `json:"description"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// APIKeys represents a collection of API keys.
type APIKeys []*APIKey

// APIKeyCreationRequest represents the request to create an API key.
type APIKeyCreationRequest struct {
	Description string `json:"description"`
}

func SetOptionalField[T any](value T) *T {
	return &value
}
