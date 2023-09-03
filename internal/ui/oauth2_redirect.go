// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/logger"
	"miniflux.app/v2/internal/oauth2"
	"miniflux.app/v2/internal/ui/session"
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

	auth := oauth2.GenerateAuthorization(authProvider.GetConfig())

	sess.SetOAuth2State(auth.State())
	sess.SetOAuth2CodeVerifier(auth.CodeVerifier())

	html.Redirect(w, r, auth.RedirectURL())
}
