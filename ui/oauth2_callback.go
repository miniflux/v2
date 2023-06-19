// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"errors"
	"net/http"

	"miniflux.app/config"
	"miniflux.app/http/cookie"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/locale"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/ui/session"
)

func (h *handler) oauth2Callback(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	printer := locale.NewPrinter(request.UserLanguage(r))
	sess := session.New(h.store, request.SessionID(r))

	provider := request.RouteStringParam(r, "provider")
	if provider == "" {
		logger.Error("[OAuth2] Invalid or missing provider")
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	code := request.QueryStringParam(r, "code", "")
	if code == "" {
		logger.Error("[OAuth2] No code received on callback")
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	state := request.QueryStringParam(r, "state", "")
	if state == "" || state != request.OAuth2State(r) {
		logger.Error(`[OAuth2] Invalid state value: got "%s" instead of "%s"`, state, request.OAuth2State(r))
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(r.Context()).FindProvider(provider)
	if err != nil {
		logger.Error("[OAuth2] %v", err)
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	profile, err := authProvider.GetProfile(r.Context(), code)
	if err != nil {
		logger.Error("[OAuth2] %v", err)
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	logger.Info("[OAuth2] [ClientIP=%s] Successful auth for %s", clientIP, profile)

	if request.IsAuthenticated(r) {
		loggedUser, err := h.store.UserByID(request.UserID(r))
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		if h.store.AnotherUserWithFieldExists(loggedUser.ID, profile.Key, profile.ID) {
			logger.Error("[OAuth2] User #%d cannot be associated because it is already associated with another user", loggedUser.ID)
			sess.NewFlashErrorMessage(printer.Printf("error.duplicate_linked_account"))
			html.Redirect(w, r, route.Path(h.router, "settings"))
			return
		}

		authProvider.PopulateUserWithProfileID(loggedUser, profile)
		if err := h.store.UpdateUser(loggedUser); err != nil {
			html.ServerError(w, r, err)
			return
		}

		sess.NewFlashMessage(printer.Printf("alert.account_linked"))
		html.Redirect(w, r, route.Path(h.router, "settings"))
		return
	}

	user, err := h.store.UserByField(profile.Key, profile.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if user == nil {
		if !config.Opts.IsOAuth2UserCreationAllowed() {
			html.Forbidden(w, r)
			return
		}

		if h.store.UserExists(profile.Username) {
			html.BadRequest(w, r, errors.New(printer.Printf("error.user_already_exists")))
			return
		}

		userCreationRequest := &model.UserCreationRequest{Username: profile.Username}
		authProvider.PopulateUserCreationWithProfileID(userCreationRequest, profile)

		user, err = h.store.CreateUser(userCreationRequest)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	}

	sessionToken, _, err := h.store.CreateUserSessionFromUsername(user.Username, r.UserAgent(), clientIP)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	logger.Info("[OAuth2] [ClientIP=%s] username=%s (%s) just logged in", clientIP, user.Username, profile)

	h.store.SetLastLogin(user.ID)
	sess.SetLanguage(user.Language)
	sess.SetTheme(user.Theme)

	http.SetCookie(w, cookie.New(
		cookie.CookieUserSessionID,
		sessionToken,
		config.Opts.HTTPS,
		config.Opts.BasePath(),
	))

	html.Redirect(w, r, route.Path(h.router, "unread"))
}
