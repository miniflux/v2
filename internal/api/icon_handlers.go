// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"errors"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

func (h *handler) getIconByFeedIDHandler(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if feedID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	icon, err := h.store.IconByFeedID(request.UserID(r), feedID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if icon == nil {
		response.JSONNotFound(w, r)
		return
	}

	response.JSON(w, r, &feedIconResponse{
		ID:       icon.ID,
		MimeType: icon.MimeType,
		Data:     icon.DataURL(),
	})
}

func (h *handler) getIconByIconIDHandler(w http.ResponseWriter, r *http.Request) {
	iconID := request.RouteInt64Param(r, "iconID")
	if iconID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid icon ID"))
		return
	}

	icon, err := h.store.IconByID(iconID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if icon == nil {
		response.JSONNotFound(w, r)
		return
	}

	response.JSON(w, r, &feedIconResponse{
		ID:       icon.ID,
		MimeType: icon.MimeType,
		Data:     icon.DataURL(),
	})
}
