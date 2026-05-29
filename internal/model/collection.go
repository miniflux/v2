// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"fmt"
	"time"
)

// Collection represents a user-defined grouping of saved entries that can be
// exported to disk or shared with an external registry.
type Collection struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Public    bool      `json:"public"`
	ItemCount int       `json:"item_count"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Collection) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, Title=%s", c.ID, c.UserID, c.Title)
}

// CollectionItem links an entry to a collection.
type CollectionItem struct {
	ID           int64  `json:"id"`
	CollectionID int64  `json:"collection_id"`
	EntryID      int64  `json:"entry_id"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	Content      string `json:"content"`
}

// Collections represents a list of collections.
type Collections []*Collection

// CollectionItems represents a list of collection items.
type CollectionItems []*CollectionItem

// CollectionCreationRequest is the payload used to create a collection.
type CollectionCreationRequest struct {
	Title  string `json:"title"`
	Public bool   `json:"public"`
}

// CollectionModificationRequest is the payload used to update a collection.
type CollectionModificationRequest struct {
	Title  *string `json:"title"`
	Public *bool   `json:"public"`
}

// Patch applies the modification request to an existing collection.
func (r *CollectionModificationRequest) Patch(collection *Collection) {
	if r.Title != nil {
		collection.Title = *r.Title
	}
	if r.Public != nil {
		collection.Public = *r.Public
	}
}
