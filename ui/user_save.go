// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/response/html"
	"miniflux.app/http/request"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) saveUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if !user.IsAdmin {
		html.Forbidden(w, r)
		return
	}

	userForm := form.NewUserForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))
	view.Set("form", userForm)

	if err := userForm.ValidateCreation(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("create_user"))
		return
	}

	if h.store.UserExists(userForm.Username) {
		view.Set("errorMessage", "error.user_already_exists")
		html.OK(w, r, view.Render("create_user"))
		return
	}

	newUser := userForm.ToUser()
	if err := h.store.CreateUser(newUser); err != nil {
		logger.Error("[Controller:SaveUser] %v", err)
		view.Set("errorMessage", "error.unable_to_create_user")
		html.OK(w, r, view.Render("create_user"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "users"))
}
