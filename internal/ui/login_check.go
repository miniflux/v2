// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/cookie"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) checkLogin(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	sess := session.New(h.store, request.SessionID(r))
	authForm := form.NewAuthForm(r)

	view := view.New(h.tpl, r, sess)
	view.Set("errorMessage", "error.bad_credentials")
	view.Set("form", authForm)

	if err := authForm.Validate(); err != nil {
		slog.Warn("Validation error during login check",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", authForm.Username),
			slog.Any("error", err),
		)
		html.OK(w, r, view.Render("login"))
		return
	}

	if err := h.store.CheckPassword(authForm.Username, authForm.Password); err != nil {
		slog.Warn("Incorrect username or password",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", authForm.Username),
			slog.Any("error", err),
		)
		html.OK(w, r, view.Render("login"))
		return
	}

	sessionToken, userID, err := h.store.CreateUserSessionFromUsername(authForm.Username, r.UserAgent(), clientIP)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	slog.Info("User authenticated successfully with username/password",
		slog.Bool("authentication_successful", true),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
		slog.String("username", authForm.Username),
	)

	h.store.SetLastLogin(userID)

	user, err := h.store.UserByID(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

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
