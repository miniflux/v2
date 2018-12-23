// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"fmt"
	"net/http"

	"miniflux.app/http/client"
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
		SiteURL:      feed.SiteURL,
		FeedURL:      feed.FeedURL,
		Title:        feed.Title,
		ScraperRules: feed.ScraperRules,
		RewriteRules: feed.RewriteRules,
		Crawler:      feed.Crawler,
		CacheMedia:   feed.CacheMedia,
		UserAgent:    feed.UserAgent,
		CategoryID:   feed.Category.ID,
		Username:     feed.Username,
		Password:     feed.Password,
	}

	all, count, size, err := h.store.MediaStatisticsByFeed(feedID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", feedForm)
	view.Set("categories", categories)
	view.Set("feed", feed)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))
	view.Set("defaultUserAgent", client.DefaultUserAgent)
	view.Set("mediaCount", all)
	view.Set("cacheCount", count)
	view.Set("cacheSize", byteSizeHumanReadable(size))

	html.OK(w, r, view.Render("edit_feed"))
}

func byteSizeHumanReadable(size int) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
		PB
		EB
		ZB
		YB
	)
	unit := ""
	sz := float64(size)
	if sz < KB {
		unit = "B"
	} else if sz < MB {
		unit = "KB"
		sz = sz / KB
	} else if sz < GB {
		unit = "MB"
		sz = sz / MB
	} else if sz < TB {
		unit = "GB"
		sz = sz / GB
	} else if sz < PB {
		unit = "TB"
		sz = sz / TB
	} else if sz < EB {
		unit = "PB"
		sz = sz / PB
	} else if sz < ZB {
		unit = "EB"
		sz = sz / EB
	} else if sz < YB {
		unit = "ZB"
		sz = sz / ZB
	} else {
		unit = "YB"
		sz = sz / YB
	}
	return fmt.Sprintf("%.2f%s", sz, unit)
}
