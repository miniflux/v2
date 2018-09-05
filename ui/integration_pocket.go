// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/response"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/integration/pocket"
	"miniflux.app/logger"
	"miniflux.app/ui/session"
)

// PocketAuthorize redirects the end-user to Pocket website to authorize the application.
func (c *Controller) PocketAuthorize(w http.ResponseWriter, r *http.Request) {
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	integration, err := c.store.Integration(user.ID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	sess := session.New(c.store, request.SessionID(r))
	connector := pocket.NewConnector(c.cfg.PocketConsumerKey(integration.PocketConsumerKey))
	redirectURL := c.cfg.BaseURL() + route.Path(c.router, "pocketCallback")
	requestToken, err := connector.RequestToken(redirectURL)
	if err != nil {
		logger.Error("[Pocket:Authorize] %v", err)
		sess.NewFlashErrorMessage(c.translator.GetLanguage(request.UserLanguage(r)).Get("Unable to fetch request token from Pocket!"))
		response.Redirect(w, r, route.Path(c.router, "integrations"))
		return
	}

	sess.SetPocketRequestToken(requestToken)
	response.Redirect(w, r, connector.AuthorizationURL(requestToken, redirectURL))
}

// PocketCallback saves the personal access token after the authorization step.
func (c *Controller) PocketCallback(w http.ResponseWriter, r *http.Request) {
	sess := session.New(c.store, request.SessionID(r))

	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	integration, err := c.store.Integration(user.ID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	connector := pocket.NewConnector(c.cfg.PocketConsumerKey(integration.PocketConsumerKey))
	accessToken, err := connector.AccessToken(request.PocketRequestToken(r))
	if err != nil {
		logger.Error("[Pocket:Callback] %v", err)
		sess.NewFlashErrorMessage(c.translator.GetLanguage(request.UserLanguage(r)).Get("Unable to fetch access token from Pocket!"))
		response.Redirect(w, r, route.Path(c.router, "integrations"))
		return
	}

	sess.SetPocketRequestToken("")
	integration.PocketAccessToken = accessToken

	err = c.store.UpdateIntegration(integration)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	sess.NewFlashMessage(c.translator.GetLanguage(request.UserLanguage(r)).Get("Your Pocket account is now linked!"))
	response.Redirect(w, r, route.Path(c.router, "integrations"))
}
