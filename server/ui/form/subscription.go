// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form

import (
	"net/http"
	"strconv"

	"github.com/miniflux/miniflux2/errors"
)

// SubscriptionForm represents the subscription form.
type SubscriptionForm struct {
	URL        string
	CategoryID int64
	Crawler    bool
}

// Validate makes sure the form values are valid.
func (s *SubscriptionForm) Validate() error {
	if s.URL == "" || s.CategoryID == 0 {
		return errors.NewLocalizedError("The URL and the category are mandatory.")
	}

	return nil
}

// NewSubscriptionForm returns a new SubscriptionForm.
func NewSubscriptionForm(r *http.Request) *SubscriptionForm {
	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil {
		categoryID = 0
	}

	return &SubscriptionForm{
		URL:        r.FormValue("url"),
		Crawler:    r.FormValue("crawler") == "1",
		CategoryID: int64(categoryID),
	}
}
