// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form

import (
	"errors"
	"net/http"
	"strconv"
)

type SubscriptionForm struct {
	URL        string
	CategoryID int64
}

func (s *SubscriptionForm) Validate() error {
	if s.URL == "" || s.CategoryID == 0 {
		return errors.New("The URL and the category are mandatory.")
	}

	return nil
}

func NewSubscriptionForm(r *http.Request) *SubscriptionForm {
	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil {
		categoryID = 0
	}

	return &SubscriptionForm{
		URL:        r.FormValue("url"),
		CategoryID: int64(categoryID),
	}
}
