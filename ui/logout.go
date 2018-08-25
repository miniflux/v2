// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/cookie"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/session"
)

// Logout destroy the session and redirects the user to the login page.
func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	sess := session.New(c.store, ctx)

	user, err := c.store.UserByID(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	sess.SetLanguage(user.Language)
	sess.SetTheme(user.Theme)

	if err := c.store.RemoveUserSessionByToken(user.ID, ctx.UserSessionToken()); err != nil {
		logger.Error("[Controller:Logout] %v", err)
	}

	http.SetCookie(w, cookie.Expired(
		cookie.CookieUserSessionID,
		c.cfg.IsHTTPS,
		c.cfg.BasePath(),
	))

	response.Redirect(w, r, route.Path(c.router, "login"))
}
