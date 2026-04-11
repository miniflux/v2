// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/view"
)

// EditUser shows the form to edit a user.
func (h *handler) showEditUserPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if !user.IsAdmin {
		response.HTMLForbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	selectedUser, err := h.store.UserByID(userID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if selectedUser == nil {
		response.HTMLNotFound(w, r)
		return
	}

	userForm := &form.UserForm{
		Username: selectedUser.Username,
		IsAdmin:  selectedUser.IsAdmin,
	}

	view := view.New(h.tpl, r)
	view.Set("form", userForm)
	view.Set("selected_user", selectedUser)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	response.HTML(w, r, view.Render("edit_user"))
}
