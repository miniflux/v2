// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if !user.IsAdmin {
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
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))
	view.Set("selected_user", selectedUser)
	view.Set("form", userForm)

	if err := userForm.ValidateModification(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("edit_user"))
		return
	}

	if h.store.AnotherUserExists(selectedUser.ID, userForm.Username) {
		view.Set("errorMessage", "error.user_already_exists")
		html.OK(w, r, view.Render("edit_user"))
		return
	}

	userForm.Merge(selectedUser)
	if err := h.store.UpdateUser(selectedUser); err != nil {
		logger.Error("[UI:UpdateUser] %v", err)
		view.Set("errorMessage", "error.unable_to_update_user")
		html.OK(w, r, view.Render("edit_user"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "users"))
}
