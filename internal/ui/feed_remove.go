// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

func (h *handler) removeFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")

	exists, err := h.store.FeedExists(request.UserID(r), feedID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}
	if !exists {
		response.HTMLNotFound(w, r)
		return
	}

	if err := h.store.RemoveFeed(request.UserID(r), feedID); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/feeds"))
}
