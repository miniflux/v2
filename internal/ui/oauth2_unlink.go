// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/locale"
)

func (h *handler) oauth2Unlink(w http.ResponseWriter, r *http.Request) {
	if config.Opts.DisableLocalAuth() {
		slog.Warn("blocking oauth2 unlink attempt, local auth is disabled",
			slog.String("user_agent", r.UserAgent()),
		)
		response.HTMLRedirect(w, r, h.routePath("/"))
		return
	}

	provider := request.RouteStringParam(r, "provider")
	if provider == "" {
		slog.Warn("Invalid or missing OAuth2 provider")
		response.HTMLRedirect(w, r, h.routePath("/"))
		return
	}

	authProvider, err := getOAuth2Manager(r.Context()).FindProvider(provider)
	if err != nil {
		slog.Error("Unable to initialize OAuth2 provider",
			slog.String("provider", provider),
			slog.Any("error", err),
		)
		response.HTMLRedirect(w, r, h.routePath("/settings"))
		return
	}

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	hasPassword, err := h.store.HasPassword(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	sess := request.WebSession(r)
	printer := locale.NewPrinter(sess.Language())
	if !hasPassword {
		sess.SetErrorMessage(printer.Print("error.unlink_account_without_password"))
		response.HTMLRedirect(w, r, h.routePath("/settings"))
		return
	}

	authProvider.UnsetUserProfileID(user)
	if err := h.store.UpdateUser(user); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	sess.SetSuccessMessage(printer.Print("alert.account_unlinked"))
	response.HTMLRedirect(w, r, h.routePath("/settings"))
}
