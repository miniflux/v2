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
)

// RefreshFeed refresh a subscription and redirect to the feed entries page.
func (c *Controller) RefreshFeed(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		html.BadRequest(w, err)
		return
	}

	ctx := context.New(r)
	if err := c.feedHandler.RefreshFeed(ctx.UserID(), feedID); err != nil {
		logger.Error("[Controller:RefreshFeed] %v", err)
	}

	response.Redirect(w, r, route.Path(c.router, "feedEntries", "feedID", feedID))
}

// RefreshAllFeeds refresh all feeds in the background for the current user.
func (c *Controller) RefreshAllFeeds(w http.ResponseWriter, r *http.Request) {
	userID := context.New(r).UserID()
	jobs, err := c.store.NewUserBatch(userID, c.store.CountFeeds(userID))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	go func() {
		c.pool.Push(jobs)
	}()

	response.Redirect(w, r, route.Path(c.router, "feeds"))
}
