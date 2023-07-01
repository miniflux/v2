// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/locale"
	"miniflux.app/logger"
	"miniflux.app/ui/session"
)

func (h *handler) oauth2Unlink(w http.ResponseWriter, r *http.Request) {
	printer := locale.NewPrinter(request.UserLanguage(r))
	provider := request.RouteStringParam(r, "provider")
	if provider == "" {
		logger.Info("[OAuth2] Invalid or missing provider")
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(r.Context()).FindProvider(provider)
	if err != nil {
		logger.Error("[OAuth2] %v", err)
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
		sess.NewFlashErrorMessage(printer.Printf("error.unlink_account_without_password"))
		html.Redirect(w, r, route.Path(h.router, "settings"))
		return
	}

	authProvider.UnsetUserProfileID(user)
	if err := h.store.UpdateUser(user); err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess.NewFlashMessage(printer.Printf("alert.account_unlinked"))
	html.Redirect(w, r, route.Path(h.router, "settings"))
}
