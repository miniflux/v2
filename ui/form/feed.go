// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"
	"strconv"

	"miniflux.app/model"
)

// FeedForm represents a feed form in the UI
type FeedForm struct {
	FeedURL                     string
	SiteURL                     string
	Title                       string
	ScraperRules                string
	RewriteRules                string
	BlocklistRules              string
	KeeplistRules               string
	Crawler                     bool
	UserAgent                   string
	Cookie                      string
	CategoryID                  int64
	Username                    string
	Password                    string
	IgnoreHTTPCache             bool
	AllowSelfSignedCertificates bool
	ApplyFilterToContent        bool
	FetchViaProxy               bool
	Disabled                    bool
}

// Merge updates the fields of the given feed.
func (f FeedForm) Merge(feed *model.Feed) *model.Feed {
	feed.Category.ID = f.CategoryID
	feed.Title = f.Title
	feed.SiteURL = f.SiteURL
	feed.FeedURL = f.FeedURL
	feed.ScraperRules = f.ScraperRules
	feed.RewriteRules = f.RewriteRules
	feed.BlocklistRules = f.BlocklistRules
	feed.KeeplistRules = f.KeeplistRules
	feed.Crawler = f.Crawler
	feed.UserAgent = f.UserAgent
	feed.Cookie = f.Cookie
	feed.ParsingErrorCount = 0
	feed.ParsingErrorMsg = ""
	feed.Username = f.Username
	feed.Password = f.Password
	feed.IgnoreHTTPCache = f.IgnoreHTTPCache
	feed.AllowSelfSignedCertificates = f.AllowSelfSignedCertificates
	feed.ApplyFilterToContent = f.ApplyFilterToContent
	feed.FetchViaProxy = f.FetchViaProxy
	feed.Disabled = f.Disabled
	return feed
}

// NewFeedForm parses the HTTP request and returns a FeedForm
func NewFeedForm(r *http.Request) *FeedForm {
	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil {
		categoryID = 0
	}
	return &FeedForm{
		FeedURL:                     r.FormValue("feed_url"),
		SiteURL:                     r.FormValue("site_url"),
		Title:                       r.FormValue("title"),
		ScraperRules:                r.FormValue("scraper_rules"),
		UserAgent:                   r.FormValue("user_agent"),
		Cookie:                      r.FormValue("cookie"),
		RewriteRules:                r.FormValue("rewrite_rules"),
		BlocklistRules:              r.FormValue("blocklist_rules"),
		KeeplistRules:               r.FormValue("keeplist_rules"),
		Crawler:                     r.FormValue("crawler") == "1",
		CategoryID:                  int64(categoryID),
		Username:                    r.FormValue("feed_username"),
		Password:                    r.FormValue("feed_password"),
		IgnoreHTTPCache:             r.FormValue("ignore_http_cache") == "1",
		AllowSelfSignedCertificates: r.FormValue("allow_self_signed_certificates") == "1",
		ApplyFilterToContent:        r.FormValue("apply_filter_to_content") == "1",
		FetchViaProxy:               r.FormValue("fetch_via_proxy") == "1",
		Disabled:                    r.FormValue("disabled") == "1",
	}
}
