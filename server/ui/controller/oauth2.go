// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"log"
	"net/http"

	"github.com/miniflux/miniflux2/config"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/oauth2"
	"github.com/tomasen/realip"
)

// OAuth2Redirect redirects the user to the consent page to ask for permission.
func (c *Controller) OAuth2Redirect(ctx *core.Context, request *core.Request, response *core.Response) {
	provider := request.StringParam("provider", "")
	if provider == "" {
		log.Println("[OAuth2] Invalid or missing provider")
		response.Redirect(ctx.Route("login"))
		return
	}

	authProvider, err := getOAuth2Manager(c.cfg).Provider(provider)
	if err != nil {
		log.Println("[OAuth2]", err)
		response.Redirect(ctx.Route("login"))
		return
	}

	response.Redirect(authProvider.GetRedirectURL(ctx.CsrfToken()))
}

// OAuth2Callback receives the authorization code and create a new session.
func (c *Controller) OAuth2Callback(ctx *core.Context, request *core.Request, response *core.Response) {
	provider := request.StringParam("provider", "")
	if provider == "" {
		log.Println("[OAuth2] Invalid or missing provider")
		response.Redirect(ctx.Route("login"))
		return
	}

	code := request.QueryStringParam("code", "")
	if code == "" {
		log.Println("[OAuth2] No code received on callback")
		response.Redirect(ctx.Route("login"))
		return
	}

	state := request.QueryStringParam("state", "")
	if state != ctx.CsrfToken() {
		log.Println("[OAuth2] Invalid state value")
		response.Redirect(ctx.Route("login"))
		return
	}

	authProvider, err := getOAuth2Manager(c.cfg).Provider(provider)
	if err != nil {
		log.Println("[OAuth2]", err)
		response.Redirect(ctx.Route("login"))
		return
	}

	profile, err := authProvider.GetProfile(code)
	if err != nil {
		log.Println("[OAuth2]", err)
		response.Redirect(ctx.Route("login"))
		return
	}

	if ctx.IsAuthenticated() {
		user := ctx.LoggedUser()
		if err := c.store.UpdateExtraField(user.ID, profile.Key, profile.ID); err != nil {
			response.HTML().ServerError(err)
			return
		}

		response.Redirect(ctx.Route("settings"))
		return
	}

	user, err := c.store.UserByExtraField(profile.Key, profile.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if user == nil {
		if c.cfg.GetInt("OAUTH2_USER_CREATION", 0) == 0 {
			response.HTML().Forbidden()
			return
		}

		user = model.NewUser()
		user.Username = profile.Username
		user.IsAdmin = false
		user.Extra[profile.Key] = profile.ID

		if err := c.store.CreateUser(user); err != nil {
			response.HTML().ServerError(err)
			return
		}
	}

	sessionToken, err := c.store.CreateSession(
		user.Username,
		request.Request().UserAgent(),
		realip.RealIP(request.Request()),
	)

	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	log.Printf("[UI:OAuth2Callback] username=%s just logged in\n", user.Username)

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

// OAuth2Unlink unlink an account from the external provider.
func (c *Controller) OAuth2Unlink(ctx *core.Context, request *core.Request, response *core.Response) {
	provider := request.StringParam("provider", "")
	if provider == "" {
		log.Println("[OAuth2] Invalid or missing provider")
		response.Redirect(ctx.Route("login"))
		return
	}

	authProvider, err := getOAuth2Manager(c.cfg).Provider(provider)
	if err != nil {
		log.Println("[OAuth2]", err)
		response.Redirect(ctx.Route("settings"))
		return
	}

	user := ctx.LoggedUser()
	if err := c.store.RemoveExtraField(user.ID, authProvider.GetUserExtraKey()); err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.Redirect(ctx.Route("settings"))
	return
}

func getOAuth2Manager(cfg *config.Config) *oauth2.Manager {
	return oauth2.NewManager(
		cfg.Get("OAUTH2_CLIENT_ID", ""),
		cfg.Get("OAUTH2_CLIENT_SECRET", ""),
		cfg.Get("OAUTH2_REDIRECT_URL", ""),
	)
}
