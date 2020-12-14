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

type feedIcon struct {
	ID       int64  `json:"id"`
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type entriesResponse struct {
	Total   int           `json:"total"`
	Entries model.Entries `json:"entries"`
}

type feedCreation struct {
	FeedURL        string `json:"feed_url"`
	CategoryID     int64  `json:"category_id"`
	UserAgent      string `json:"user_agent"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Crawler        bool   `json:"crawler"`
	FetchViaProxy  bool   `json:"fetch_via_proxy"`
	ScraperRules   string `json:"scraper_rules"`
	RewriteRules   string `json:"rewrite_rules"`
	BlocklistRules string `json:"blocklist_rules"`
	KeeplistRules  string `json:"keeplist_rules"`
}

type subscriptionDiscovery struct {
	URL           string `json:"url"`
	UserAgent     string `json:"user_agent"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	FetchViaProxy bool   `json:"fetch_via_proxy"`
}

type feedModification struct {
	FeedURL        *string `json:"feed_url"`
	SiteURL        *string `json:"site_url"`
	Title          *string `json:"title"`
	ScraperRules   *string `json:"scraper_rules"`
	RewriteRules   *string `json:"rewrite_rules"`
	BlocklistRules *string `json:"blocklist_rules"`
	KeeplistRules  *string `json:"keeplist_rules"`
	Crawler        *bool   `json:"crawler"`
	UserAgent      *string `json:"user_agent"`
	Username       *string `json:"username"`
	Password       *string `json:"password"`
	CategoryID     *int64  `json:"category_id"`
	Disabled       *bool   `json:"disabled"`
}

func (f *feedModification) Update(feed *model.Feed) {
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
}

type userModification struct {
	Username       *string `json:"username"`
	Password       *string `json:"password"`
	IsAdmin        *bool   `json:"is_admin"`
	Theme          *string `json:"theme"`
	Language       *string `json:"language"`
	Timezone       *string `json:"timezone"`
	EntryDirection *string `json:"entry_sorting_direction"`
	EntriesPerPage *int    `json:"entries_per_page"`
}

func (u *userModification) Update(user *model.User) {
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

	if u.EntriesPerPage != nil {
		user.EntriesPerPage = *u.EntriesPerPage
	}
}

func decodeUserModificationPayload(r io.ReadCloser) (*userModification, error) {
	defer r.Close()

	var user userModification
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("Unable to decode user modification JSON object: %v", err)
	}

	return &user, nil
}

func decodeUserCreationPayload(r io.ReadCloser) (*model.User, error) {
	defer r.Close()

	var user model.User
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("Unable to decode user modification JSON object: %v", err)
	}

	return &user, nil
}

func decodeURLPayload(r io.ReadCloser) (*subscriptionDiscovery, error) {
	defer r.Close()

	var s subscriptionDiscovery
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&s); err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %v", err)
	}

	return &s, nil
}

func decodeEntryStatusPayload(r io.ReadCloser) ([]int64, string, error) {
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

func decodeFeedCreationPayload(r io.ReadCloser) (*feedCreation, error) {
	defer r.Close()

	var fc feedCreation
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&fc); err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %v", err)
	}

	return &fc, nil
}

func decodeFeedModificationPayload(r io.ReadCloser) (*feedModification, error) {
	defer r.Close()

	var feed feedModification
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("Unable to decode feed modification JSON object: %v", err)
	}

	return &feed, nil
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
