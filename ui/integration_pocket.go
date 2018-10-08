// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/response/html"
	"miniflux.app/http/request"
	"miniflux.app/http/route"
	"miniflux.app/integration/pocket"
	"miniflux.app/locale"
	"miniflux.app/logger"
	"miniflux.app/ui/session"
)

// PocketAuthorize redirects the end-user to Pocket website to authorize the application.
func (c *Controller) PocketAuthorize(w http.ResponseWriter, r *http.Request) {
	printer := locale.NewPrinter(request.UserLanguage(r))
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	integration, err := c.store.Integration(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(c.store, request.SessionID(r))
	connector := pocket.NewConnector(c.cfg.PocketConsumerKey(integration.PocketConsumerKey))
	redirectURL := c.cfg.BaseURL() + route.Path(c.router, "pocketCallback")
	requestToken, err := connector.RequestToken(redirectURL)
	if err != nil {
		logger.Error("[Pocket:Authorize] %v", err)
		sess.NewFlashErrorMessage(printer.Printf("error.pocket_request_token"))
		html.Redirect(w, r, route.Path(c.router, "integrations"))
		return
	}

	sess.SetPocketRequestToken(requestToken)
	html.Redirect(w, r, connector.AuthorizationURL(requestToken, redirectURL))
}

// PocketCallback saves the personal access token after the authorization step.
func (c *Controller) PocketCallback(w http.ResponseWriter, r *http.Request) {
	printer := locale.NewPrinter(request.UserLanguage(r))
	sess := session.New(c.store, request.SessionID(r))

	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	integration, err := c.store.Integration(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	connector := pocket.NewConnector(c.cfg.PocketConsumerKey(integration.PocketConsumerKey))
	accessToken, err := connector.AccessToken(request.PocketRequestToken(r))
	if err != nil {
		logger.Error("[Pocket:Callback] %v", err)
		sess.NewFlashErrorMessage(printer.Printf("error.pocket_access_token"))
		html.Redirect(w, r, route.Path(c.router, "integrations"))
		return
	}

	sess.SetPocketRequestToken("")
	integration.PocketAccessToken = accessToken

	err = c.store.UpdateIntegration(integration)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess.NewFlashMessage(printer.Printf("alert.pocket_linked"))
	html.Redirect(w, r, route.Path(c.router, "integrations"))
}
