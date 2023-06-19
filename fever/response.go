// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever // import "miniflux.app/fever"

import (
	"time"
)

type baseResponse struct {
	Version       int   `json:"api_version"`
	Authenticated int   `json:"auth"`
	LastRefresh   int64 `json:"last_refreshed_on_time"`
}

func (b *baseResponse) SetCommonValues() {
	b.Version = 3
	b.Authenticated = 1
	b.LastRefresh = time.Now().Unix()
}

/*
The default response is a JSON object containing two members:

	api_version contains the version of the API responding (positive integer)
	auth whether the request was successfully authenticated (boolean integer)

The API can also return XML by passing xml as the optional value of the api argument like so:

http://yourdomain.com/fever/?api=xml

The top level XML element is named response.

The response to each successfully authenticated request will have auth set to 1 and include
at least one additional member:

	last_refreshed_on_time contains the time of the most recently refreshed (not updated)
	feed (Unix timestamp/integer)
*/
func newBaseResponse() baseResponse {
	r := baseResponse{}
	r.SetCommonValues()
	return r
}

func newAuthFailureResponse() baseResponse {
	return baseResponse{Version: 3, Authenticated: 0}
}

type groupsResponse struct {
	baseResponse
	Groups      []group       `json:"groups"`
	FeedsGroups []feedsGroups `json:"feeds_groups"`
}

type feedsResponse struct {
	baseResponse
	Feeds       []feed        `json:"feeds"`
	FeedsGroups []feedsGroups `json:"feeds_groups"`
}

type faviconsResponse struct {
	baseResponse
	Favicons []favicon `json:"favicons"`
}

type itemsResponse struct {
	baseResponse
	Items []item `json:"items"`
	Total int    `json:"total_items"`
}

type unreadResponse struct {
	baseResponse
	ItemIDs string `json:"unread_item_ids"`
}

type savedResponse struct {
	baseResponse
	ItemIDs string `json:"saved_item_ids"`
}

type group struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type feedsGroups struct {
	GroupID int64  `json:"group_id"`
	FeedIDs string `json:"feed_ids"`
}

type feed struct {
	ID          int64  `json:"id"`
	FaviconID   int64  `json:"favicon_id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	SiteURL     string `json:"site_url"`
	IsSpark     int    `json:"is_spark"`
	LastUpdated int64  `json:"last_updated_on_time"`
}

type item struct {
	ID        int64  `json:"id"`
	FeedID    int64  `json:"feed_id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	HTML      string `json:"html"`
	URL       string `json:"url"`
	IsSaved   int    `json:"is_saved"`
	IsRead    int    `json:"is_read"`
	CreatedAt int64  `json:"created_on_time"`
}

type favicon struct {
	ID   int64  `json:"id"`
	Data string `json:"data"`
}
