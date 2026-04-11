// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"errors"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/urllib"
)

func (h *handler) checkLogin(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	view := view.New(h.tpl, r)
	redirectURL := r.FormValue("redirect_url")
	view.Set("redirectURL", redirectURL)

	if config.Opts.DisableLocalAuth() {
		slog.Warn("blocking local auth login attempt, local auth is disabled",
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
		)
		response.HTML(w, r, view.Render("login"))
		return
	}

	authForm := form.NewAuthForm(r)
	view.Set("errorMessage", locale.NewLocalizedError("error.bad_credentials").Translate(request.WebSession(r).Language()))
	view.Set("form", authForm)

	if validationErr := authForm.Validate(); validationErr != nil {
		translatedErrorMessage := validationErr.Translate(request.WebSession(r).Language())
		slog.Warn("Validation error during login check",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", authForm.Username),
			slog.Any("error", translatedErrorMessage),
		)
		response.HTML(w, r, view.Render("login"))
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
		response.HTML(w, r, view.Render("login"))
		return
	}

	user, err := h.store.UserByUsername(authForm.Username)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}
	if user == nil {
		response.HTMLServerError(w, r, errors.New("authenticated user not found"))
		return
	}

	slog.Info("User authenticated successfully with username/password",
		slog.Bool("authentication_successful", true),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", user.ID),
		slog.String("username", authForm.Username),
	)

	h.store.SetLastLogin(user.ID)
	if err := authenticateWebSession(w, r, h.store, user); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if redirectURL != "" && urllib.IsRelativePath(redirectURL) {
		response.HTMLRedirect(w, r, redirectURL)
		return
	}

	response.HTMLRedirect(w, r, h.basePath+"/"+user.DefaultHomePage)
}
