// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
)

func (h *handler) removeCategoryFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	categoryID := request.RouteInt64Param(r, "categoryID")

	if !h.store.CategoryFeedExists(request.UserID(r), categoryID, feedID) {
		html.NotFound(w, r)
		return
	}

	if err := h.store.RemoveFeed(request.UserID(r), feedID); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "categoryFeeds", "categoryID", categoryID))
}
