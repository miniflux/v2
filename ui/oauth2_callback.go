// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/cookie"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/http/route"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/ui/session"
)

// OAuth2Callback receives the authorization code and create a new session.
func (c *Controller) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	sess := session.New(c.store, ctx)

	provider := request.Param(r, "provider", "")
	if provider == "" {
		logger.Error("[OAuth2] Invalid or missing provider")
		response.Redirect(w, r, route.Path(c.router, "login"))
		return
	}

	code := request.QueryParam(r, "code", "")
	if code == "" {
		logger.Error("[OAuth2] No code received on callback")
		response.Redirect(w, r, route.Path(c.router, "login"))
		return
	}

	state := request.QueryParam(r, "state", "")
	if state == "" || state != ctx.OAuth2State() {
		logger.Error(`[OAuth2] Invalid state value: got "%s" instead of "%s"`, state, ctx.OAuth2State())
		response.Redirect(w, r, route.Path(c.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(c.cfg).Provider(provider)
	if err != nil {
		logger.Error("[OAuth2] %v", err)
		response.Redirect(w, r, route.Path(c.router, "login"))
		return
	}

	profile, err := authProvider.GetProfile(code)
	if err != nil {
		logger.Error("[OAuth2] %v", err)
		response.Redirect(w, r, route.Path(c.router, "login"))
		return
	}

	if ctx.IsAuthenticated() {
		user, err := c.store.UserByExtraField(profile.Key, profile.ID)
		if err != nil {
			html.ServerError(w, err)
			return
		}

		if user != nil {
			logger.Error("[OAuth2] User #%d cannot be associated because %s is already associated", ctx.UserID(), user.Username)
			sess.NewFlashErrorMessage(c.translator.GetLanguage(ctx.UserLanguage()).Get("There is already someone associated with this provider!"))
			response.Redirect(w, r, route.Path(c.router, "settings"))
			return
		}

		if err := c.store.UpdateExtraField(ctx.UserID(), profile.Key, profile.ID); err != nil {
			html.ServerError(w, err)
			return
		}

		sess.NewFlashMessage(c.translator.GetLanguage(ctx.UserLanguage()).Get("Your external account is now linked!"))
		response.Redirect(w, r, route.Path(c.router, "settings"))
		return
	}

	user, err := c.store.UserByExtraField(profile.Key, profile.ID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	if user == nil {
		if !c.cfg.IsOAuth2UserCreationAllowed() {
			html.Forbidden(w)
			return
		}

		user = model.NewUser()
		user.Username = profile.Username
		user.IsAdmin = false
		user.Extra[profile.Key] = profile.ID

		if err := c.store.CreateUser(user); err != nil {
			html.ServerError(w, err)
			return
		}
	}

	sessionToken, _, err := c.store.CreateUserSession(user.Username, r.UserAgent(), request.RealIP(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	logger.Info("[Controller:OAuth2Callback] username=%s just logged in", user.Username)
	c.store.SetLastLogin(user.ID)
	sess.SetLanguage(user.Language)
	sess.SetTheme(user.Theme)

	http.SetCookie(w, cookie.New(
		cookie.CookieUserSessionID,
		sessionToken,
		c.cfg.IsHTTPS,
		c.cfg.BasePath(),
	))

	response.Redirect(w, r, route.Path(c.router, "unread"))
}
