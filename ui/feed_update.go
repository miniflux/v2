// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/http/route"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/ui/form"
	"github.com/miniflux/miniflux/ui/session"
	"github.com/miniflux/miniflux/ui/view"
)

// UpdateFeed update a subscription and redirect to the feed entries page.
func (c *Controller) UpdateFeed(w http.ResponseWriter, r *http.Request) {
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

	feedForm := form.NewFeedForm(r)

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("form", feedForm)
	view.Set("categories", categories)
	view.Set("feed", feed)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))

	if err := feedForm.ValidateModification(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, view.Render("edit_feed"))
		return
	}

	err = c.store.UpdateFeed(feedForm.Merge(feed))
	if err != nil {
		logger.Error("[Controller:EditFeed] %v", err)
		view.Set("errorMessage", "Unable to update this feed.")
		html.OK(w, view.Render("edit_feed"))
		return
	}

	response.Redirect(w, r, route.Path(c.router, "feedEntries", "feedID", feed.ID))
}
