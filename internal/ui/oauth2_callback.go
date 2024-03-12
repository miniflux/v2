// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/cookie"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/session"
)

func (h *handler) oauth2Callback(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	printer := locale.NewPrinter(request.UserLanguage(r))
	sess := session.New(h.store, request.SessionID(r))

	provider := request.RouteStringParam(r, "provider")
	if provider == "" {
		slog.Warn("Invalid or missing OAuth2 provider")
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	code := request.QueryStringParam(r, "code", "")
	if code == "" {
		slog.Warn("No code received on OAuth2 callback")
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	state := request.QueryStringParam(r, "state", "")
	if subtle.ConstantTimeCompare([]byte(state), []byte(request.OAuth2State(r))) == 0 {
		slog.Warn("Invalid OAuth2 state value received",
			slog.String("expected", request.OAuth2State(r)),
			slog.String("received", state),
		)
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(r.Context()).FindProvider(provider)
	if err != nil {
		slog.Error("Unable to initialize OAuth2 provider",
			slog.String("provider", provider),
			slog.Any("error", err),
		)
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	profile, err := authProvider.GetProfile(r.Context(), code, request.OAuth2CodeVerifier(r))
	if err != nil {
		slog.Warn("Unable to get OAuth2 profile from provider",
			slog.String("provider", provider),
			slog.Any("error", err),
		)
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	if request.IsAuthenticated(r) {
		loggedUser, err := h.store.UserByID(request.UserID(r))
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		if h.store.AnotherUserWithFieldExists(loggedUser.ID, profile.Key, profile.ID) {
			slog.Error("Oauth2 user cannot be associated because it is already associated with another user",
				slog.Int64("user_id", loggedUser.ID),
				slog.String("oauth2_provider", provider),
				slog.String("oauth2_profile_id", profile.ID),
			)
			sess.NewFlashErrorMessage(printer.Print("error.duplicate_linked_account"))
			html.Redirect(w, r, route.Path(h.router, "settings"))
			return
		}

		authProvider.PopulateUserWithProfileID(loggedUser, profile)
		if err := h.store.UpdateUser(loggedUser); err != nil {
			html.ServerError(w, r, err)
			return
		}

		sess.NewFlashMessage(printer.Print("alert.account_linked"))
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
			html.BadRequest(w, r, errors.New(printer.Print("error.user_already_exists")))
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

	slog.Info("User authenticated successfully using OAuth2",
		slog.Bool("authentication_successful", true),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Username),
	)

	h.store.SetLastLogin(user.ID)
	sess.SetLanguage(user.Language)
	sess.SetTheme(user.Theme)

	http.SetCookie(w, cookie.New(
		cookie.CookieUserSessionID,
		sessionToken,
		config.Opts.HTTPS,
		config.Opts.BasePath(),
	))

	html.Redirect(w, r, route.Path(h.router, user.DefaultHomePage))
}
