// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"net/http"
	"time"
)

func (h *handler) markFeedAsRead(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	userID := request.UserID(r)

	feed, err := h.store.FeedByID(userID, feedID)

	before := feed.CheckedAt

	if request.HasQueryParam(r, "before") {
		before = time.Unix(request.QueryInt64Param(r, "before", 0), 0)
	}

	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if feed == nil {
		html.NotFound(w, r)
		return
	}

	if err = h.store.MarkFeedAsRead(userID, feedID, before); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "feeds"))
}
