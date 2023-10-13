// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
)

func (h *handler) getIconByFeedID(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")

	if !h.store.HasIcon(feedID) {
		json.NotFound(w, r)
		return
	}

	icon, err := h.store.IconByFeedID(request.UserID(r), feedID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if icon == nil {
		json.NotFound(w, r)
		return
	}

	json.OK(w, r, &feedIconResponse{
		ID:       icon.ID,
		MimeType: icon.MimeType,
		Data:     icon.DataURL(),
	})
}

func (h *handler) getIconByIconID(w http.ResponseWriter, r *http.Request) {
	iconID := request.RouteInt64Param(r, "iconID")

	icon, err := h.store.IconByID(iconID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if icon == nil {
		json.NotFound(w, r)
		return
	}

	json.OK(w, r, &feedIconResponse{
		ID:       icon.ID,
		MimeType: icon.MimeType,
		Data:     icon.DataURL(),
	})
}
