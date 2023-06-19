// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/session"
)

func (h *handler) oauth2Redirect(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))

	provider := request.RouteStringParam(r, "provider")
	if provider == "" {
		logger.Error("[OAuth2] Invalid or missing provider: %s", provider)
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(r.Context()).FindProvider(provider)
	if err != nil {
		logger.Error("[OAuth2] %v", err)
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	html.Redirect(w, r, authProvider.GetRedirectURL(sess.NewOAuth2State()))
}
