// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/client"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// EditFeed shows the form to modify a subscription.
func (c *Controller) EditFeed(w http.ResponseWriter, r *http.Request) {
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	feedID := request.RouteInt64Param(r, "feedID")
	feed, err := c.store.FeedByID(user.ID, feedID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	if feed == nil {
		html.NotFound(w)
		return
	}

	categories, err := c.store.Categories(user.ID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	feedForm := form.FeedForm{
		SiteURL:      feed.SiteURL,
		FeedURL:      feed.FeedURL,
		Title:        feed.Title,
		ScraperRules: feed.ScraperRules,
		RewriteRules: feed.RewriteRules,
		Crawler:      feed.Crawler,
		UserAgent:    feed.UserAgent,
		CategoryID:   feed.Category.ID,
		Username:     feed.Username,
		Password:     feed.Password,
	}

	sess := session.New(c.store, request.SessionID(r))
	view := view.New(c.tpl, r, sess)
	view.Set("form", feedForm)
	view.Set("categories", categories)
	view.Set("feed", feed)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))
	view.Set("defaultUserAgent", client.DefaultUserAgent)

	html.OK(w, r, view.Render("edit_feed"))
}
