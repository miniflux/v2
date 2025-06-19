// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
	"strconv"

	"miniflux.app/v2/internal/model"
)

// FeedForm represents a feed form in the UI
type FeedForm struct {
	FeedURL                     string
	SiteURL                     string
	Title                       string
	Description                 string
	ScraperRules                string
	RewriteRules                string
	UrlRewriteRules             string
	BlocklistRules              string
	KeeplistRules               string
	BlockFilterEntryRules       string
	KeepFilterEntryRules        string
	Crawler                     bool
	UserAgent                   string
	Cookie                      string
	CategoryID                  int64
	Username                    string
	Password                    string
	IgnoreHTTPCache             bool
	AllowSelfSignedCertificates bool
	FetchViaProxy               bool
	Disabled                    bool
	NoMediaPlayer               bool
	HideGlobally                bool
	CategoryHidden              bool // Category has "hide_globally"
	AppriseServiceURLs          string
	WebhookURL                  string
	DisableHTTP2                bool
	NtfyEnabled                 bool
	NtfyPriority                int
	NtfyTopic                   string
	PushoverEnabled             bool
	PushoverPriority            int
	ProxyURL                    string
}

// Merge updates the fields of the given feed.
func (f FeedForm) Merge(feed *model.Feed) *model.Feed {
	feed.Category.ID = f.CategoryID
	feed.Title = f.Title
	feed.SiteURL = f.SiteURL
	feed.FeedURL = f.FeedURL
	feed.Description = f.Description
	feed.ScraperRules = f.ScraperRules
	feed.RewriteRules = f.RewriteRules
	feed.UrlRewriteRules = f.UrlRewriteRules
	feed.BlocklistRules = f.BlocklistRules
	feed.KeeplistRules = f.KeeplistRules
	feed.BlockFilterEntryRules = f.BlockFilterEntryRules
	feed.KeepFilterEntryRules = f.KeepFilterEntryRules
	feed.Crawler = f.Crawler
	feed.UserAgent = f.UserAgent
	feed.Cookie = f.Cookie
	feed.ParsingErrorCount = 0
	feed.ParsingErrorMsg = ""
	feed.Username = f.Username
	feed.Password = f.Password
	feed.IgnoreHTTPCache = f.IgnoreHTTPCache
	feed.AllowSelfSignedCertificates = f.AllowSelfSignedCertificates
	feed.FetchViaProxy = f.FetchViaProxy
	feed.Disabled = f.Disabled
	feed.NoMediaPlayer = f.NoMediaPlayer
	feed.HideGlobally = f.HideGlobally
	feed.AppriseServiceURLs = f.AppriseServiceURLs
	feed.WebhookURL = f.WebhookURL
	feed.DisableHTTP2 = f.DisableHTTP2
	feed.NtfyEnabled = f.NtfyEnabled
	feed.NtfyPriority = f.NtfyPriority
	feed.NtfyTopic = f.NtfyTopic
	feed.PushoverEnabled = f.PushoverEnabled
	feed.PushoverPriority = f.PushoverPriority
	feed.ProxyURL = f.ProxyURL
	return feed
}

// NewFeedForm parses the HTTP request and returns a FeedForm
func NewFeedForm(r *http.Request) *FeedForm {
	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil {
		categoryID = 0
	}

	ntfyPriority, err := strconv.Atoi(r.FormValue("ntfy_priority"))
	if err != nil {
		ntfyPriority = 0
	}

	pushoverPriority, err := strconv.Atoi(r.FormValue("pushover_priority"))
	if err != nil {
		pushoverPriority = 0
	}

	return &FeedForm{
		FeedURL:                     r.FormValue("feed_url"),
		SiteURL:                     r.FormValue("site_url"),
		Title:                       r.FormValue("title"),
		Description:                 r.FormValue("description"),
		ScraperRules:                r.FormValue("scraper_rules"),
		UserAgent:                   r.FormValue("user_agent"),
		Cookie:                      r.FormValue("cookie"),
		RewriteRules:                r.FormValue("rewrite_rules"),
		UrlRewriteRules:             r.FormValue("urlrewrite_rules"),
		BlocklistRules:              r.FormValue("blocklist_rules"),
		KeeplistRules:               r.FormValue("keeplist_rules"),
		BlockFilterEntryRules:       r.FormValue("block_filter_entry_rules"),
		KeepFilterEntryRules:        r.FormValue("keep_filter_entry_rules"),
		Crawler:                     r.FormValue("crawler") == "1",
		CategoryID:                  int64(categoryID),
		Username:                    r.FormValue("feed_username"),
		Password:                    r.FormValue("feed_password"),
		IgnoreHTTPCache:             r.FormValue("ignore_http_cache") == "1",
		AllowSelfSignedCertificates: r.FormValue("allow_self_signed_certificates") == "1",
		FetchViaProxy:               r.FormValue("fetch_via_proxy") == "1",
		Disabled:                    r.FormValue("disabled") == "1",
		NoMediaPlayer:               r.FormValue("no_media_player") == "1",
		HideGlobally:                r.FormValue("hide_globally") == "1",
		AppriseServiceURLs:          r.FormValue("apprise_service_urls"),
		WebhookURL:                  r.FormValue("webhook_url"),
		DisableHTTP2:                r.FormValue("disable_http2") == "1",
		NtfyEnabled:                 r.FormValue("ntfy_enabled") == "1",
		NtfyPriority:                ntfyPriority,
		NtfyTopic:                   r.FormValue("ntfy_topic"),
		PushoverEnabled:             r.FormValue("pushover_enabled") == "1",
		PushoverPriority:            pushoverPriority,
		ProxyURL:                    r.FormValue("proxy_url"),
	}
}
