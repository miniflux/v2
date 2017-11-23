// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/form"

	"github.com/tomasen/realip"
)

// ShowLoginPage shows the login form.
func (c *Controller) ShowLoginPage(ctx *core.Context, request *core.Request, response *core.Response) {
	if ctx.IsAuthenticated() {
		response.Redirect(ctx.Route("unread"))
		return
	}

	response.HTML().Render("login", tplParams{
		"csrf": ctx.CsrfToken(),
	})
}

// CheckLogin validates the username/password and redirects the user to the unread page.
func (c *Controller) CheckLogin(ctx *core.Context, request *core.Request, response *core.Response) {
	authForm := form.NewAuthForm(request.Request())
	tplParams := tplParams{
		"errorMessage": "Invalid username or password.",
		"csrf":         ctx.CsrfToken(),
	}

	if err := authForm.Validate(); err != nil {
		log.Println(err)
		response.HTML().Render("login", tplParams)
		return
	}

	if err := c.store.CheckPassword(authForm.Username, authForm.Password); err != nil {
		log.Println(err)
		response.HTML().Render("login", tplParams)
		return
	}

	sessionToken, err := c.store.CreateSession(
		authForm.Username,
		request.Request().UserAgent(),
		realip.RealIP(request.Request()),
	)

	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	log.Printf("[UI:CheckLogin] username=%s just logged in\n", authForm.Username)

	cookie := &http.Cookie{
		Name:     "sessionID",
		Value:    sessionToken,
		Path:     "/",
		Secure:   request.IsHTTPS(),
		HttpOnly: true,
	}

	response.SetCookie(cookie)
	response.Redirect(ctx.Route("unread"))
}

// Logout destroy the session and redirects the user to the login page.
func (c *Controller) Logout(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	sessionCookie := request.Cookie("sessionID")
	if err := c.store.RemoveSessionByToken(user.ID, sessionCookie); err != nil {
		log.Printf("[UI:Logout] %v", err)
	}

	cookie := &http.Cookie{
		Name:     "sessionID",
		Value:    "",
		Path:     "/",
		Secure:   request.IsHTTPS(),
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	response.SetCookie(cookie)
	response.Redirect(ctx.Route("login"))
}
