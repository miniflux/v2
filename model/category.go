// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import (
	"errors"
	"fmt"
)

// Category represents a category in the system.
type Category struct {
	ID        int64  `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	UserID    int64  `json:"user_id,omitempty"`
	FeedCount int    `json:"nb_feeds,omitempty"`
}

func (c *Category) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, Title=%s", c.ID, c.UserID, c.Title)
}

// ValidateCategoryCreation validates a category during the creation.
func (c Category) ValidateCategoryCreation() error {
	if c.Title == "" {
		return errors.New("The title is mandatory")
	}

	if c.UserID == 0 {
		return errors.New("The userID is mandatory")
	}

	return nil
}

// ValidateCategoryModification validates a category during the modification.
func (c Category) ValidateCategoryModification() error {
	if c.Title == "" {
		return errors.New("The title is mandatory")
	}

	if c.UserID == 0 {
		return errors.New("The userID is mandatory")
	}

	if c.ID <= 0 {
		return errors.New("The ID is mandatory")
	}

	return nil
}

// Categories represents a list of categories.
type Categories []*Category
