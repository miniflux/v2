// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/config"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
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
		ScraperRules:                feed.ScraperRules,
		RewriteRules:                feed.RewriteRules,
		BlocklistRules:              feed.BlocklistRules,
		KeeplistRules:               feed.KeeplistRules,
		Crawler:                     feed.Crawler,
		UserAgent:                   feed.UserAgent,
		Cookie:                      feed.Cookie,
		CategoryID:                  feed.Category.ID,
		Username:                    feed.Username,
		Password:                    feed.Password,
		IgnoreHTTPCache:             feed.IgnoreHTTPCache,
		AllowSelfSignedCertificates: feed.AllowSelfSignedCertificates,
		ApplyFilterToContent:        feed.ApplyFilterToContent,
		FetchViaProxy:               feed.FetchViaProxy,
		Disabled:                    feed.Disabled,
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
	view.Set("hasProxyConfigured", config.Opts.HasHTTPClientProxyConfigured())

	html.OK(w, r, view.Render("edit_feed"))
}
