// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/model"

import (
	"time"
)

// Entry statuses and default sorting order.
const (
	EntryStatusUnread       = "unread"
	EntryStatusRead         = "read"
	EntryStatusRemoved      = "removed"
	DefaultSortingOrder     = "published_at"
	DefaultSortingDirection = "asc"
)

// Entry represents a feed item in the system.
type Entry struct {
	ID          int64         `json:"id"`
	UserID      int64         `json:"user_id"`
	FeedID      int64         `json:"feed_id"`
	Status      string        `json:"status"`
	Hash        string        `json:"hash"`
	Title       string        `json:"title"`
	URL         string        `json:"url"`
	CommentsURL string        `json:"comments_url"`
	Date        time.Time     `json:"published_at"`
	CreatedAt   time.Time     `json:"created_at"`
	ChangedAt   time.Time     `json:"changed_at"`
	Content     string        `json:"content"`
	Author      string        `json:"author"`
	ShareCode   string        `json:"share_code"`
	Starred     bool          `json:"starred"`
	ReadingTime int           `json:"reading_time"`
	Enclosures  EnclosureList `json:"enclosures"`
	Feed        *Feed         `json:"feed,omitempty"`
	Tags        []string      `json:"tags"`
}

// Entries represents a list of entries.
type Entries []*Entry

// EntriesStatusUpdateRequest represents a request to change entries status.
type EntriesStatusUpdateRequest struct {
	EntryIDs []int64 `json:"entry_ids"`
	Status   string  `json:"status"`
}
