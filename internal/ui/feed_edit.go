// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showEditFeedPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	feedID := request.RouteInt64Param(r, "feedID")
	feed, err := h.store.FeedByID(user.ID, feedID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if feed == nil {
		html.NotFound(w, r)
		return
	}

	categories, err := h.store.Categories(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	feedForm := form.FeedForm{
		SiteURL:                     feed.SiteURL,
		FeedURL:                     feed.FeedURL,
		Title:                       feed.Title,
		Description:                 feed.Description,
		ScraperRules:                feed.ScraperRules,
		RewriteRules:                feed.RewriteRules,
		UrlRewriteRules:             feed.UrlRewriteRules,
		BlocklistRules:              feed.BlocklistRules,
		KeeplistRules:               feed.KeeplistRules,
		BlockFilterEntryRules:       feed.BlockFilterEntryRules,
		KeepFilterEntryRules:        feed.KeepFilterEntryRules,
		Crawler:                     feed.Crawler,
		UserAgent:                   feed.UserAgent,
		Cookie:                      feed.Cookie,
		CategoryID:                  feed.Category.ID,
		Username:                    feed.Username,
		Password:                    feed.Password,
		IgnoreHTTPCache:             feed.IgnoreHTTPCache,
		AllowSelfSignedCertificates: feed.AllowSelfSignedCertificates,
		FetchViaProxy:               feed.FetchViaProxy,
		Disabled:                    feed.Disabled,
		NoMediaPlayer:               feed.NoMediaPlayer,
		HideGlobally:                feed.HideGlobally,
		CategoryHidden:              feed.Category.HideGlobally,
		AppriseServiceURLs:          feed.AppriseServiceURLs,
		WebhookURL:                  feed.WebhookURL,
		DisableHTTP2:                feed.DisableHTTP2,
		NtfyEnabled:                 feed.NtfyEnabled,
		NtfyPriority:                feed.NtfyPriority,
		NtfyTopic:                   feed.NtfyTopic,
		PushoverEnabled:             feed.PushoverEnabled,
		PushoverPriority:            feed.PushoverPriority,
		ProxyURL:                    feed.ProxyURL,
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", feedForm)
	view.Set("categories", categories)
	view.Set("feed", feed)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())
	view.Set("hasProxyConfigured", config.Opts.HasHTTPClientProxyURLConfigured())

	html.OK(w, r, view.Render("edit_feed"))
}
