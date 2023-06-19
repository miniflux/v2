// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/model"

import "fmt"

// Category represents a feed category.
type Category struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	UserID       int64  `json:"user_id"`
	HideGlobally bool   `json:"hide_globally"`
	FeedCount    *int   `json:"feed_count,omitempty"`
	TotalUnread  *int   `json:"total_unread,omitempty"`
}

func (c *Category) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, Title=%s", c.ID, c.UserID, c.Title)
}

// CategoryRequest represents the request to create or update a category.
type CategoryRequest struct {
	Title        string `json:"title"`
	HideGlobally string `json:"hide_globally"`
}

// Patch updates category fields.
func (cr *CategoryRequest) Patch(category *Category) {
	category.Title = cr.Title
	category.HideGlobally = cr.HideGlobally != ""
}

// Categories represents a list of categories.
type Categories []*Category
