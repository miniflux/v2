// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/oauth2"
)

func (h *handler) oauth2Redirect(w http.ResponseWriter, r *http.Request) {
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
		response.HTMLRedirect(w, r, h.routePath("/"))
		return
	}

	auth := oauth2.GenerateAuthorization(authProvider.Config())

	request.WebSession(r).StartOAuth2Flow(auth.State(), auth.CodeVerifier())

	response.HTMLRedirect(w, r, auth.RedirectURL())
}
