// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import "fmt"

// Category represents a feed category.
type Category struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	UserID       int64  `json:"user_id"`
	HideGlobally bool   `json:"hide_globally"`
	FeedCount    int    `json:"-"`
	TotalUnread  int    `json:"-"`
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
