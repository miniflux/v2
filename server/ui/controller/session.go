// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/server/core"
)

// ShowSessions shows the list of active sessions.
func (c *Controller) ShowSessions(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	sessions, err := c.store.UserSessions(user.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	sessionCookie := request.Cookie("sessionID")
	response.HTML().Render("sessions", args.Merge(tplParams{
		"sessions":            sessions,
		"currentSessionToken": sessionCookie,
		"menu":                "settings",
	}))
}

// RemoveSession remove a session.
func (c *Controller) RemoveSession(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	sessionID, err := request.IntegerParam("sessionID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	err = c.store.RemoveUserSessionByID(user.ID, sessionID)
	if err != nil {
		logger.Error("[Controller:RemoveSession] %v", err)
	}

	response.Redirect(ctx.Route("sessions"))
}
