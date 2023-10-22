// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if !loggedUser.IsAdmin {
		html.Forbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	selectedUser, err := h.store.UserByID(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if selectedUser == nil {
		html.NotFound(w, r)
		return
	}

	userForm := form.NewUserForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("menu", "settings")
	view.Set("user", loggedUser)
	view.Set("countUnread", h.store.CountUnreadEntries(loggedUser.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID))
	view.Set("selected_user", selectedUser)
	view.Set("form", userForm)

	if validationErr := userForm.ValidateModification(); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(loggedUser.Language))
		html.OK(w, r, view.Render("edit_user"))
		return
	}

	if h.store.AnotherUserExists(selectedUser.ID, userForm.Username) {
		view.Set("errorMessage", locale.NewLocalizedError("error.user_already_exists").Translate(loggedUser.Language))
		html.OK(w, r, view.Render("edit_user"))
		return
	}

	userForm.Merge(selectedUser)
	if err := h.store.UpdateUser(selectedUser); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "users"))
}
