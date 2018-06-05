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
)

// RemoveUser deletes a user from the database.
func (c *Controller) RemoveUser(w http.ResponseWriter, r *http.Request) {
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

	if err := c.store.RemoveUser(selectedUser.ID); err != nil {
		html.ServerError(w, err)
		return
	}

	response.Redirect(w, r, route.Path(c.router, "users"))
}
