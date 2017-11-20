// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/form"
	"log"
	"net/http"
	"time"

	"github.com/tomasen/realip"
)

func (c *Controller) ShowLoginPage(ctx *core.Context, request *core.Request, response *core.Response) {
	if ctx.IsAuthenticated() {
		response.Redirect(ctx.GetRoute("unread"))
		return
	}

	response.Html().Render("login", tplParams{
		"csrf": ctx.GetCsrfToken(),
	})
}

func (c *Controller) CheckLogin(ctx *core.Context, request *core.Request, response *core.Response) {
	authForm := form.NewAuthForm(request.GetRequest())
	tplParams := tplParams{
		"errorMessage": "Invalid username or password.",
		"csrf":         ctx.GetCsrfToken(),
	}

	if err := authForm.Validate(); err != nil {
		log.Println(err)
		response.Html().Render("login", tplParams)
		return
	}

	if err := c.store.CheckPassword(authForm.Username, authForm.Password); err != nil {
		log.Println(err)
		response.Html().Render("login", tplParams)
		return
	}

	sessionToken, err := c.store.CreateSession(
		authForm.Username,
		request.GetHeaders().Get("User-Agent"),
		realip.RealIP(request.GetRequest()),
	)
	if err != nil {
		response.Html().ServerError(err)
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
	response.Redirect(ctx.GetRoute("unread"))
}

func (c *Controller) Logout(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	sessionCookie := request.GetCookie("sessionID")
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
	response.Redirect(ctx.GetRoute("login"))
}
