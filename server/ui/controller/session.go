// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/miniflux/miniflux2/server/core"
	"log"
)

func (c *Controller) ShowSessions(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	sessions, err := c.store.GetSessions(user.ID)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	sessionCookie := request.GetCookie("sessionID")
	response.Html().Render("sessions", args.Merge(tplParams{
		"sessions":            sessions,
		"currentSessionToken": sessionCookie,
		"menu":                "settings",
	}))
}

func (c *Controller) RemoveSession(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	sessionID, err := request.GetIntegerParam("sessionID")
	if err != nil {
		response.Html().BadRequest(err)
		return
	}

	err = c.store.RemoveSessionByID(user.ID, sessionID)
	if err != nil {
		log.Println("[UI:RemoveSession]", err)
	}

	response.Redirect(ctx.GetRoute("sessions"))
}
