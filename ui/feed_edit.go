// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/ui/form"
	"github.com/miniflux/miniflux/ui/session"
	"github.com/miniflux/miniflux/ui/view"
)

// EditFeed shows the form to modify a subscription.
func (c *Controller) EditFeed(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)

	user, err := c.store.UserByID(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		html.BadRequest(w, err)
		return
	}

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
		CategoryID:   feed.Category.ID,
		Username:     feed.Username,
		Password:     feed.Password,
	}

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("form", feedForm)
	view.Set("categories", categories)
	view.Set("feed", feed)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))

	html.OK(w, view.Render("edit_feed"))
}
