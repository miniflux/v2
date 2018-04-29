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
)

// RemoveFeed deletes a subscription from the database and redirect to the list of feeds page.
func (c *Controller) RemoveFeed(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		html.ServerError(w, err)
		return
	}

	ctx := context.New(r)
	if err := c.store.RemoveFeed(ctx.UserID(), feedID); err != nil {
		html.ServerError(w, err)
		return
	}

	response.Redirect(w, r, route.Path(c.router, "feeds"))
}
