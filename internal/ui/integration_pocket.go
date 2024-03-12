// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/integration/pocket"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/ui/session"
)

func (h *handler) pocketAuthorize(w http.ResponseWriter, r *http.Request) {
	printer := locale.NewPrinter(request.UserLanguage(r))
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	integration, err := h.store.Integration(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	connector := pocket.NewConnector(config.Opts.PocketConsumerKey(integration.PocketConsumerKey))
	redirectURL := config.Opts.RootURL() + route.Path(h.router, "pocketCallback")
	requestToken, err := connector.RequestToken(redirectURL)
	if err != nil {
		slog.Warn("Pocket authorization request failed",
			slog.Any("user_id", user.ID),
			slog.Any("error", err),
		)
		sess.NewFlashErrorMessage(printer.Print("error.pocket_request_token"))
		html.Redirect(w, r, route.Path(h.router, "integrations"))
		return
	}

	sess.SetPocketRequestToken(requestToken)
	html.Redirect(w, r, connector.AuthorizationURL(requestToken, redirectURL))
}

func (h *handler) pocketCallback(w http.ResponseWriter, r *http.Request) {
	printer := locale.NewPrinter(request.UserLanguage(r))
	sess := session.New(h.store, request.SessionID(r))

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	integration, err := h.store.Integration(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	connector := pocket.NewConnector(config.Opts.PocketConsumerKey(integration.PocketConsumerKey))
	accessToken, err := connector.AccessToken(request.PocketRequestToken(r))
	if err != nil {
		slog.Warn("Unable to get Pocket access token",
			slog.Any("user_id", user.ID),
			slog.Any("error", err),
		)
		sess.NewFlashErrorMessage(printer.Print("error.pocket_access_token"))
		html.Redirect(w, r, route.Path(h.router, "integrations"))
		return
	}

	sess.SetPocketRequestToken("")
	integration.PocketAccessToken = accessToken

	err = h.store.UpdateIntegration(integration)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess.NewFlashMessage(printer.Print("alert.pocket_linked"))
	html.Redirect(w, r, route.Path(h.router, "integrations"))
}
