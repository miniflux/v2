// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"net/url"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

func (h *handler) markTagAsRead(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	tagName, err := url.PathUnescape(request.RouteStringParam(r, "tagName"))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if !h.store.TagExists(userID, tagName) {
		response.HTMLNotFound(w, r)
		return
	}

	if err = h.store.MarkTagAsRead(userID, tagName, time.Now()); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/tags"))
}
