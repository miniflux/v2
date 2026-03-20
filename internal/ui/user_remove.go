// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"errors"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

func (h *handler) removeUser(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if !loggedUser.IsAdmin {
		response.HTMLForbidden(w, r)
		return
	}

	selectedUserID := request.RouteInt64Param(r, "userID")
	selectedUser, err := h.store.UserByID(selectedUserID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if selectedUser == nil {
		response.HTMLNotFound(w, r)
		return
	}

	if selectedUser.ID == loggedUser.ID {
		response.HTMLBadRequest(w, r, errors.New("you cannot remove yourself"))
		return
	}

	if err := h.store.RemoveUser(selectedUser.ID); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/users"))
}
