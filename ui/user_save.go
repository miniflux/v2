// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// SaveUser validate and save the new user into the database.
func (c *Controller) SaveUser(w http.ResponseWriter, r *http.Request) {
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

	userForm := form.NewUserForm(r)

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("form", userForm)

	if err := userForm.ValidateCreation(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("create_user"))
		return
	}

	if c.store.UserExists(userForm.Username) {
		view.Set("errorMessage", "This user already exists.")
		html.OK(w, r, view.Render("create_user"))
		return
	}

	newUser := userForm.ToUser()
	if err := c.store.CreateUser(newUser); err != nil {
		logger.Error("[Controller:SaveUser] %v", err)
		view.Set("errorMessage", "Unable to create this user.")
		html.OK(w, r, view.Render("create_user"))
		return
	}

	response.Redirect(w, r, route.Path(c.router, "users"))
}
