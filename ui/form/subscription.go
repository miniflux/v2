// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"
	"strconv"

	"miniflux.app/errors"
)

// SubscriptionForm represents the subscription form.
type SubscriptionForm struct {
	URL            string
	CategoryID     int64
	Crawler        bool
	FetchViaProxy  bool
	UserAgent      string
	Username       string
	Password       string
	ScraperRules   string
	RewriteRules   string
	BlocklistRules string
	KeeplistRules  string
}

// Validate makes sure the form values are valid.
func (s *SubscriptionForm) Validate() error {
	if s.URL == "" || s.CategoryID == 0 {
		return errors.NewLocalizedError("error.feed_mandatory_fields")
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
		URL:            r.FormValue("url"),
		CategoryID:     int64(categoryID),
		Crawler:        r.FormValue("crawler") == "1",
		FetchViaProxy:  r.FormValue("fetch_via_proxy") == "1",
		UserAgent:      r.FormValue("user_agent"),
		Username:       r.FormValue("feed_username"),
		Password:       r.FormValue("feed_password"),
		ScraperRules:   r.FormValue("scraper_rules"),
		RewriteRules:   r.FormValue("rewrite_rules"),
		BlocklistRules: r.FormValue("blocklist_rules"),
		KeeplistRules:  r.FormValue("keeplist_rules"),
	}
}
