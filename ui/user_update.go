// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/http/route"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/ui/form"
	"github.com/miniflux/miniflux/ui/session"
	"github.com/miniflux/miniflux/ui/view"
)

// UpdateUser validate and update a user.
func (c *Controller) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)

	user, err := c.store.UserByID(ctx.UserID())
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

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
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
