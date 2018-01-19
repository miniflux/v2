// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"github.com/miniflux/miniflux/http/cookie"
	"github.com/miniflux/miniflux/http/handler"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/ui/form"

	"github.com/tomasen/realip"
)

// ShowLoginPage shows the login form.
func (c *Controller) ShowLoginPage(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	if ctx.IsAuthenticated() {
		response.Redirect(ctx.Route("unread"))
		return
	}

	response.HTML().Render("login", tplParams{
		"csrf": ctx.CSRF(),
	})
}

// CheckLogin validates the username/password and redirects the user to the unread page.
func (c *Controller) CheckLogin(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	authForm := form.NewAuthForm(request.Request())
	tplParams := tplParams{
		"errorMessage": "Invalid username or password.",
		"csrf":         ctx.CSRF(),
		"form":         authForm,
	}

	if err := authForm.Validate(); err != nil {
		logger.Error("[Controller:CheckLogin] %v", err)
		response.HTML().Render("login", tplParams)
		return
	}

	if err := c.store.CheckPassword(authForm.Username, authForm.Password); err != nil {
		logger.Error("[Controller:CheckLogin] %v", err)
		response.HTML().Render("login", tplParams)
		return
	}

	sessionToken, err := c.store.CreateUserSession(
		authForm.Username,
		request.Request().UserAgent(),
		realip.RealIP(request.Request()),
	)

	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	logger.Info("[Controller:CheckLogin] username=%s just logged in", authForm.Username)

	response.SetCookie(cookie.New(cookie.CookieUserSessionID, sessionToken, c.cfg.IsHTTPS))
	response.Redirect(ctx.Route("unread"))
}

// Logout destroy the session and redirects the user to the login page.
func (c *Controller) Logout(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	if err := c.store.UpdateSessionField(ctx.SessionID(), "language", user.Language); err != nil {
		logger.Error("[Controller:Logout] %v", err)
	}

	if err := c.store.RemoveUserSessionByToken(user.ID, ctx.UserSessionToken()); err != nil {
		logger.Error("[Controller:Logout] %v", err)
	}

	response.SetCookie(cookie.Expired(cookie.CookieUserSessionID, c.cfg.IsHTTPS))
	response.Redirect(ctx.Route("login"))
}
