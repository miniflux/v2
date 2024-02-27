// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/ui/session"
)

func (h *handler) oauth2Unlink(w http.ResponseWriter, r *http.Request) {
	printer := locale.NewPrinter(request.UserLanguage(r))
	provider := request.RouteStringParam(r, "provider")
	if provider == "" {
		slog.Warn("Invalid or missing OAuth2 provider")
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(r.Context()).FindProvider(provider)
	if err != nil {
		slog.Error("Unable to initialize OAuth2 provider",
			slog.String("provider", provider),
			slog.Any("error", err),
		)
		html.Redirect(w, r, route.Path(h.router, "settings"))
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	hasPassword, err := h.store.HasPassword(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if !hasPassword {
		sess.NewFlashErrorMessage(printer.Print("error.unlink_account_without_password"))
		html.Redirect(w, r, route.Path(h.router, "settings"))
		return
	}

	authProvider.UnsetUserProfileID(user)
	if err := h.store.UpdateUser(user); err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess.NewFlashMessage(printer.Print("alert.account_unlinked"))
	html.Redirect(w, r, route.Path(h.router, "settings"))
}
