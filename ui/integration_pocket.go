// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/config"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/integration/pocket"
	"miniflux.app/locale"
	"miniflux.app/logger"
	"miniflux.app/ui/session"
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
	redirectURL := config.Opts.BaseURL() + route.Path(h.router, "pocketCallback")
	requestToken, err := connector.RequestToken(redirectURL)
	if err != nil {
		logger.Error("[Pocket:Authorize] %v", err)
		sess.NewFlashErrorMessage(printer.Printf("error.pocket_request_token"))
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
		logger.Error("[Pocket:Callback] %v", err)
		sess.NewFlashErrorMessage(printer.Printf("error.pocket_access_token"))
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

	sess.NewFlashMessage(printer.Printf("alert.pocket_linked"))
	html.Redirect(w, r, route.Path(h.router, "integrations"))
}
