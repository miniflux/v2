// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

func (h *handler) markCategoryAsRead(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	categoryID := request.RouteInt64Param(r, "categoryID")

	category, err := h.store.Category(userID, categoryID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if category == nil {
		response.HTMLNotFound(w, r)
		return
	}

	if err = h.store.MarkCategoryAsRead(userID, categoryID, time.Now()); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/categories"))
}
