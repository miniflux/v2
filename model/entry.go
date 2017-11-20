// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
	"time"
)

const (
	EntryStatusUnread       = "unread"
	EntryStatusRead         = "read"
	EntryStatusRemoved      = "removed"
	DefaultSortingOrder     = "published_at"
	DefaultSortingDirection = "desc"
)

type Entry struct {
	ID         int64         `json:"id"`
	UserID     int64         `json:"user_id"`
	FeedID     int64         `json:"feed_id"`
	Status     string        `json:"status"`
	Hash       string        `json:"hash"`
	Title      string        `json:"title"`
	URL        string        `json:"url"`
	Date       time.Time     `json:"published_at"`
	Content    string        `json:"content"`
	Author     string        `json:"author"`
	Enclosures EnclosureList `json:"enclosures,omitempty"`
	Feed       *Feed         `json:"feed,omitempty"`
	Category   *Category     `json:"category,omitempty"`
}

type Entries []*Entry

func ValidateEntryStatus(status string) error {
	switch status {
	case EntryStatusRead, EntryStatusUnread, EntryStatusRemoved:
		return nil
	}

	return fmt.Errorf(`Invalid entry status, valid status values are: "%s", "%s" and "%s"`, EntryStatusRead, EntryStatusUnread, EntryStatusRemoved)
}

func ValidateEntryOrder(order string) error {
	switch order {
	case "id", "status", "published_at", "category_title", "category_id":
		return nil
	}

	return fmt.Errorf(`Invalid entry order, valid order values are: "id", "status", "published_at", "category_title", "category_id"`)
}

func ValidateDirection(direction string) error {
	switch direction {
	case "asc", "desc":
		return nil
	}

	return fmt.Errorf(`Invalid direction, valid direction values are: "asc" or "desc"`)
}

func GetOppositeDirection(direction string) string {
	if direction == "asc" {
		return "desc"
	}

	return "asc"
}
