// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Entry statuses and default sorting order.
const (
	EntryStatusUnread       = "unread"
	EntryStatusRead         = "read"
	DefaultSortingOrder     = "published_at"
	DefaultSortingDirection = "asc"
)

// MaxEntryLimit is the maximum allowed value for the "limit" query parameter
// and for the user "entries_per_page" preference.
const MaxEntryLimit = 1000

// MaxEntryIDsLimit is the maximum allowed value for the "limit" query parameter
// for the entry ID list endpoints.
const MaxEntryIDsLimit = 10000

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
	Language    string        `json:"language"`
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

	// SourceFingerprint identifies the version of the entry published by the feed.
	SourceFingerprint string `json:"-"`
}

func NewEntry() *Entry {
	return &Entry{
		Enclosures: make(EnclosureList, 0),
		Tags:       make([]string, 0),
		Feed: &Feed{
			Category: &Category{},
			Icon:     &FeedIcon{},
		},
	}
}

// ComputeSourceFingerprint hashes the fields that a refresh can overwrite.
// Call it on a parsed entry before applying local processing. The publication
// date, reading time, and entry identity do not indicate changes to these fields.
func (e *Entry) ComputeSourceFingerprint() string {
	digest := sha256.New()
	var fieldLength [8]byte
	writeField := func(value string) {
		binary.LittleEndian.PutUint64(fieldLength[:], uint64(len(value)))
		digest.Write(fieldLength[:])
		digest.Write([]byte(value))
	}

	for _, value := range []string{
		e.Title,
		e.Content,
		e.URL,
		e.CommentsURL,
		e.Author,
		e.Language,
	} {
		writeField(value)
	}
	writeField(strconv.Itoa(len(e.Tags)))
	for _, tag := range e.Tags {
		writeField(tag)
	}

	// Refresh adds or removes enclosures by URL. It keeps existing enclosures
	// with the same URL, including their metadata and playback progress.
	enclosureURLs := make([]string, 0, len(e.Enclosures))
	for _, enclosure := range e.Enclosures {
		if url := strings.TrimSpace(enclosure.URL); url != "" {
			enclosureURLs = append(enclosureURLs, url)
		}
	}
	slices.Sort(enclosureURLs)
	for _, url := range slices.Compact(enclosureURLs) {
		writeField(url)
	}

	return hex.EncodeToString(digest.Sum(nil))
}

// ShouldMarkAsReadOnView Return whether the entry should be marked as viewed considering all user settings and entry state.
func (e *Entry) ShouldMarkAsReadOnView(user *User) bool {
	// Already read, no need to mark as read again. Removed entries are not marked as read
	if e.Status != EntryStatusUnread {
		return false
	}

	// There is an enclosure, markAsRead will happen at enclosure completion time, no need to mark as read on view
	if user.MarkReadOnMediaPlayerCompletion && e.Enclosures.ContainsAudioOrVideo() {
		return false
	}

	// The user wants to mark as read on view
	return user.MarkReadOnView
}

// Entries represents a list of entries.
type Entries []*Entry

// EntriesStatusUpdateRequest represents a request to change entries status.
type EntriesStatusUpdateRequest struct {
	EntryIDs []int64 `json:"entry_ids"`
	Status   string  `json:"status"`
	Starred  *bool   `json:"starred"`
}

// EntryUpdateRequest represents a request to update an entry.
type EntryUpdateRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

func (e *EntryUpdateRequest) Patch(entry *Entry) {
	if e.Title != nil && *e.Title != "" {
		entry.Title = *e.Title
	}

	if e.Content != nil && *e.Content != "" {
		entry.Content = *e.Content
	}
}
