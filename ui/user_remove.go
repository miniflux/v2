// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"errors"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
)

func (h *handler) removeUser(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if !loggedUser.IsAdmin {
		html.Forbidden(w, r)
		return
	}

	selectedUserID := request.RouteInt64Param(r, "userID")
	selectedUser, err := h.store.UserByID(selectedUserID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if selectedUser == nil {
		html.NotFound(w, r)
		return
	}

	if selectedUser.ID == loggedUser.ID {
		html.BadRequest(w, r, errors.New("You cannot remove yourself"))
		return
	}

	if err := h.store.RemoveUser(selectedUser.ID); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "users"))
}
