// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"github.com/miniflux/miniflux/http/handler"
	"github.com/miniflux/miniflux/logger"
)

// ShowSessions shows the list of active user sessions.
func (c *Controller) ShowSessions(ctx *handler.Context, request *handler.Request, response *handler.Response) {
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

	response.HTML().Render("sessions", args.Merge(tplParams{
		"sessions":            sessions,
		"currentSessionToken": ctx.UserSessionToken(),
		"menu":                "settings",
	}))
}

// RemoveSession remove a user session.
func (c *Controller) RemoveSession(ctx *handler.Context, request *handler.Request, response *handler.Response) {
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
