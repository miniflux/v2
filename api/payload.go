// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"encoding/json"
	"fmt"
	"io"

	"miniflux.app/model"
)

type feedIconResponse struct {
	ID       int64  `json:"id"`
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type entriesResponse struct {
	Total   int           `json:"total"`
	Entries model.Entries `json:"entries"`
}

type subscriptionDiscoveryRequest struct {
	URL           string `json:"url"`
	UserAgent     string `json:"user_agent"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	FetchViaProxy bool   `json:"fetch_via_proxy"`
}

func decodeSubscriptionDiscoveryRequest(r io.ReadCloser) (*subscriptionDiscoveryRequest, error) {
	defer r.Close()

	var s subscriptionDiscoveryRequest
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&s); err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %v", err)
	}

	return &s, nil
}

type feedCreationResponse struct {
	FeedID int64 `json:"feed_id"`
}

type feedCreationRequest struct {
	FeedURL         string `json:"feed_url"`
	CategoryID      int64  `json:"category_id"`
	UserAgent       string `json:"user_agent"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Crawler         bool   `json:"crawler"`
	Disabled        bool   `json:"disabled"`
	IgnoreHTTPCache bool   `json:"ignore_http_cache"`
	FetchViaProxy   bool   `json:"fetch_via_proxy"`
	ScraperRules    string `json:"scraper_rules"`
	RewriteRules    string `json:"rewrite_rules"`
	BlocklistRules  string `json:"blocklist_rules"`
	KeeplistRules   string `json:"keeplist_rules"`
}

func decodeFeedCreationRequest(r io.ReadCloser) (*feedCreationRequest, error) {
	defer r.Close()

	var fc feedCreationRequest
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&fc); err != nil {
		return nil, fmt.Errorf("Invalid JSON payload: %v", err)
	}

	return &fc, nil
}

type feedModificationRequest struct {
	FeedURL         *string `json:"feed_url"`
	SiteURL         *string `json:"site_url"`
	Title           *string `json:"title"`
	ScraperRules    *string `json:"scraper_rules"`
	RewriteRules    *string `json:"rewrite_rules"`
	BlocklistRules  *string `json:"blocklist_rules"`
	KeeplistRules   *string `json:"keeplist_rules"`
	Crawler         *bool   `json:"crawler"`
	UserAgent       *string `json:"user_agent"`
	Username        *string `json:"username"`
	Password        *string `json:"password"`
	CategoryID      *int64  `json:"category_id"`
	Disabled        *bool   `json:"disabled"`
	IgnoreHTTPCache *bool   `json:"ignore_http_cache"`
	FetchViaProxy   *bool   `json:"fetch_via_proxy"`
}

func (f *feedModificationRequest) Update(feed *model.Feed) {
	if f.FeedURL != nil && *f.FeedURL != "" {
		feed.FeedURL = *f.FeedURL
	}

	if f.SiteURL != nil && *f.SiteURL != "" {
		feed.SiteURL = *f.SiteURL
	}

	if f.Title != nil && *f.Title != "" {
		feed.Title = *f.Title
	}

	if f.ScraperRules != nil {
		feed.ScraperRules = *f.ScraperRules
	}

	if f.RewriteRules != nil {
		feed.RewriteRules = *f.RewriteRules
	}

	if f.KeeplistRules != nil {
		feed.KeeplistRules = *f.KeeplistRules
	}

	if f.BlocklistRules != nil {
		feed.BlocklistRules = *f.BlocklistRules
	}

	if f.Crawler != nil {
		feed.Crawler = *f.Crawler
	}

	if f.UserAgent != nil {
		feed.UserAgent = *f.UserAgent
	}

	if f.Username != nil {
		feed.Username = *f.Username
	}

	if f.Password != nil {
		feed.Password = *f.Password
	}

	if f.CategoryID != nil && *f.CategoryID > 0 {
		feed.Category.ID = *f.CategoryID
	}

	if f.Disabled != nil {
		feed.Disabled = *f.Disabled
	}

	if f.IgnoreHTTPCache != nil {
		feed.IgnoreHTTPCache = *f.IgnoreHTTPCache
	}

	if f.FetchViaProxy != nil {
		feed.FetchViaProxy = *f.FetchViaProxy
	}
}

func decodeFeedModificationRequest(r io.ReadCloser) (*feedModificationRequest, error) {
	defer r.Close()

	var feed feedModificationRequest
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("Unable to decode feed modification JSON object: %v", err)
	}

	return &feed, nil
}

type userCreationRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	IsAdmin         bool   `json:"is_admin"`
	GoogleID        string `json:"google_id"`
	OpenIDConnectID string `json:"openid_connect_id"`
}

func decodeUserCreationRequest(r io.ReadCloser) (*userCreationRequest, error) {
	defer r.Close()

	var request userCreationRequest
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&request); err != nil {
		return nil, fmt.Errorf("Unable to decode user creation JSON object: %v", err)
	}

	return &request, nil
}

type userModificationRequest struct {
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
}

func (u *userModificationRequest) Update(user *model.User) {
	if u.Username != nil {
		user.Username = *u.Username
	}

	if u.Password != nil {
		user.Password = *u.Password
	}

	if u.IsAdmin != nil {
		user.IsAdmin = *u.IsAdmin
	}

	if u.Theme != nil {
		user.Theme = *u.Theme
	}

	if u.Language != nil {
		user.Language = *u.Language
	}

	if u.Timezone != nil {
		user.Timezone = *u.Timezone
	}

	if u.EntryDirection != nil {
		user.EntryDirection = *u.EntryDirection
	}

	if u.Stylesheet != nil {
		user.Stylesheet = *u.Stylesheet
	}

	if u.GoogleID != nil {
		user.GoogleID = *u.GoogleID
	}

	if u.OpenIDConnectID != nil {
		user.OpenIDConnectID = *u.OpenIDConnectID
	}

	if u.EntriesPerPage != nil {
		user.EntriesPerPage = *u.EntriesPerPage
	}

	if u.KeyboardShortcuts != nil {
		user.KeyboardShortcuts = *u.KeyboardShortcuts
	}

	if u.ShowReadingTime != nil {
		user.ShowReadingTime = *u.ShowReadingTime
	}

	if u.EntrySwipe != nil {
		user.EntrySwipe = *u.EntrySwipe
	}
}

func decodeUserModificationRequest(r io.ReadCloser) (*userModificationRequest, error) {
	defer r.Close()

	var request userModificationRequest
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&request); err != nil {
		return nil, fmt.Errorf("Unable to decode user modification JSON object: %v", err)
	}

	return &request, nil
}

func decodeEntryStatusRequest(r io.ReadCloser) ([]int64, string, error) {
	type payload struct {
		EntryIDs []int64 `json:"entry_ids"`
		Status   string  `json:"status"`
	}

	var p payload
	decoder := json.NewDecoder(r)
	defer r.Close()
	if err := decoder.Decode(&p); err != nil {
		return nil, "", fmt.Errorf("invalid JSON payload: %v", err)
	}

	return p.EntryIDs, p.Status, nil
}

type categoryRequest struct {
	Title string `json:"title"`
}

func decodeCategoryRequest(r io.ReadCloser) (*categoryRequest, error) {
	var payload categoryRequest

	decoder := json.NewDecoder(r)
	defer r.Close()
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("Unable to decode JSON object: %v", err)
	}

	return &payload, nil
}
