// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import "fmt"

// Category represents a feed category.
type Category struct {
	ID                int64  `json:"id"`
	Title             string `json:"title"`
	UserID            int64  `json:"user_id"`
	HideGlobally      bool   `json:"hide_globally"`
	Public            bool   `json:"public"`
	ShowOnHomepage    bool   `json:"show_on_homepage"`
	IsHomepageDefault bool   `json:"is_homepage_default"`
	FeedCount         *int   `json:"feed_count,omitempty"`
	TotalUnread       *int   `json:"total_unread,omitempty"`
}

func (c *Category) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, Title=%s, Public=%v, ShowOnHomepage=%v", c.ID, c.UserID, c.Title, c.Public, c.ShowOnHomepage)
}

// CategoryRequest represents the request to create or update a category.
type CategoryRequest struct {
	Title             string `json:"title"`
	HideGlobally      string `json:"hide_globally"`
	Public            string `json:"public"`
	ShowOnHomepage    string `json:"show_on_homepage"`
	IsHomepageDefault string `json:"is_homepage_default"`
}

// Patch updates category fields.
func (cr *CategoryRequest) Patch(category *Category) {
	category.Title = cr.Title
	category.HideGlobally = cr.HideGlobally != ""
	category.Public = cr.Public != ""
	category.ShowOnHomepage = cr.ShowOnHomepage != ""
	category.IsHomepageDefault = cr.IsHomepageDefault != ""
}

// Categories represents a list of categories.
type Categories []*Category
