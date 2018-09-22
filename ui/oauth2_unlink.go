// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/session"
)

// OAuth2Unlink unlink an account from the external provider.
func (c *Controller) OAuth2Unlink(w http.ResponseWriter, r *http.Request) {
	provider := request.Param(r, "provider", "")
	if provider == "" {
		logger.Info("[OAuth2] Invalid or missing provider")
		response.Redirect(w, r, route.Path(c.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(c.cfg).Provider(provider)
	if err != nil {
		logger.Error("[OAuth2] %v", err)
		response.Redirect(w, r, route.Path(c.router, "settings"))
		return
	}

	sess := session.New(c.store, request.SessionID(r))

	hasPassword, err := c.store.HasPassword(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	if !hasPassword {
		sess.NewFlashErrorMessage(c.translator.GetLanguage(request.UserLanguage(r)).Get("error.unlink_account_without_password"))
		response.Redirect(w, r, route.Path(c.router, "settings"))
		return
	}

	if err := c.store.RemoveExtraField(request.UserID(r), authProvider.GetUserExtraKey()); err != nil {
		html.ServerError(w, err)
		return
	}

	sess.NewFlashMessage(c.translator.GetLanguage(request.UserLanguage(r)).Get("alert.account_unlinked"))
	response.Redirect(w, r, route.Path(c.router, "settings"))
}
