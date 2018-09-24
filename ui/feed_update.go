// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/client"
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// UpdateFeed update a subscription and redirect to the feed entries page.
func (c *Controller) UpdateFeed(w http.ResponseWriter, r *http.Request) {
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

	feedForm := form.NewFeedForm(r)

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

	if err := feedForm.ValidateModification(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("edit_feed"))
		return
	}

	err = c.store.UpdateFeed(feedForm.Merge(feed))
	if err != nil {
		logger.Error("[Controller:EditFeed] %v", err)
		view.Set("errorMessage", "error.unable_to_update_feed")
		html.OK(w, r, view.Render("edit_feed"))
		return
	}

	response.Redirect(w, r, route.Path(c.router, "feedEntries", "feedID", feed.ID))
}
