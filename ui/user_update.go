// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// UpdateUser validate and update a user.
func (c *Controller) UpdateUser(w http.ResponseWriter, r *http.Request) {
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	if !user.IsAdmin {
		html.Forbidden(w)
		return
	}

	userID, err := request.IntParam(r, "userID")
	if err != nil {
		html.BadRequest(w, err)
		return
	}

	selectedUser, err := c.store.UserByID(userID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	if selectedUser == nil {
		html.NotFound(w)
		return
	}

	userForm := form.NewUserForm(r)

	sess := session.New(c.store, request.SessionID(r))
	view := view.New(c.tpl, r, sess)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))
	view.Set("selected_user", selectedUser)
	view.Set("form", userForm)

	if err := userForm.ValidateModification(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("edit_user"))
		return
	}

	if c.store.AnotherUserExists(selectedUser.ID, userForm.Username) {
		view.Set("errorMessage", "This user already exists.")
		html.OK(w, r, view.Render("edit_user"))
		return
	}

	userForm.Merge(selectedUser)
	if err := c.store.UpdateUser(selectedUser); err != nil {
		logger.Error("[Controller:UpdateUser] %v", err)
		view.Set("errorMessage", "Unable to update this user.")
		html.OK(w, r, view.Render("edit_user"))
		return
	}

	response.Redirect(w, r, route.Path(c.router, "users"))
}
