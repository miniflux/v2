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
)

// RemoveSession remove a user session.
func (c *Controller) RemoveSession(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)

	sessionID, err := request.IntParam(r, "sessionID")
	if err != nil {
		html.BadRequest(w, err)
		return
	}

	err = c.store.RemoveUserSessionByID(ctx.UserID(), sessionID)
	if err != nil {
		logger.Error("[Controller:RemoveSession] %v", err)
	}

	response.Redirect(w, r, route.Path(c.router, "sessions"))
}
