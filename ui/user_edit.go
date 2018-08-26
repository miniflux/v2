// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// EditUser shows the form to edit a user.
func (c *Controller) EditUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)

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

	userForm := &form.UserForm{
		Username: selectedUser.Username,
		IsAdmin:  selectedUser.IsAdmin,
	}

	view.Set("form", userForm)
	view.Set("selected_user", selectedUser)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))

	html.OK(w, r, view.Render("edit_user"))
}
