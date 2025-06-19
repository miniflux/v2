// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
	"strconv"

	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/validator"
)

// SubscriptionForm represents the subscription form.
type SubscriptionForm struct {
	URL                         string
	CategoryID                  int64
	Crawler                     bool
	FetchViaProxy               bool
	AllowSelfSignedCertificates bool
	UserAgent                   string
	Cookie                      string
	Username                    string
	Password                    string
	ScraperRules                string
	RewriteRules                string
	UrlRewriteRules             string
	BlocklistRules              string
	KeeplistRules               string
	BlockFilterEntryRules       string
	KeepFilterEntryRules        string
	DisableHTTP2                bool
	ProxyURL                    string
}

// Validate makes sure the form values locale.are valid.
func (s *SubscriptionForm) Validate() *locale.LocalizedError {
	if s.URL == "" || s.CategoryID == 0 {
		return locale.NewLocalizedError("error.feed_mandatory_fields")
	}

	if !validator.IsValidURL(s.URL) {
		return locale.NewLocalizedError("error.invalid_feed_url")
	}

	if !validator.IsValidRegex(s.BlocklistRules) {
		return locale.NewLocalizedError("error.feed_invalid_blocklist_rule")
	}

	if !validator.IsValidRegex(s.KeeplistRules) {
		return locale.NewLocalizedError("error.feed_invalid_keeplist_rule")
	}

	if !validator.IsValidRegex(s.UrlRewriteRules) {
		return locale.NewLocalizedError("error.feed_invalid_urlrewrite_rule")
	}

	if s.ProxyURL != "" && !validator.IsValidURL(s.ProxyURL) {
		return locale.NewLocalizedError("error.invalid_feed_proxy_url")
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
		URL:                         r.FormValue("url"),
		CategoryID:                  int64(categoryID),
		Crawler:                     r.FormValue("crawler") == "1",
		AllowSelfSignedCertificates: r.FormValue("allow_self_signed_certificates") == "1",
		FetchViaProxy:               r.FormValue("fetch_via_proxy") == "1",
		UserAgent:                   r.FormValue("user_agent"),
		Cookie:                      r.FormValue("cookie"),
		Username:                    r.FormValue("feed_username"),
		Password:                    r.FormValue("feed_password"),
		ScraperRules:                r.FormValue("scraper_rules"),
		RewriteRules:                r.FormValue("rewrite_rules"),
		UrlRewriteRules:             r.FormValue("urlrewrite_rules"),
		BlocklistRules:              r.FormValue("blocklist_rules"),
		KeeplistRules:               r.FormValue("keeplist_rules"),
		KeepFilterEntryRules:        r.FormValue("keep_filter_entry_rules"),
		BlockFilterEntryRules:       r.FormValue("block_filter_entry_rules"),
		DisableHTTP2:                r.FormValue("disable_http2") == "1",
		ProxyURL:                    r.FormValue("proxy_url"),
	}
}
