// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "influxeed-engine/v2/internal/ui"

import (
	"net/http"

	"influxeed-engine/v2/internal/config"
	"influxeed-engine/v2/internal/http/cookie"
	"influxeed-engine/v2/internal/http/request"
	"influxeed-engine/v2/internal/http/response/html"
	"influxeed-engine/v2/internal/http/route"
	"influxeed-engine/v2/internal/ui/session"
)

func (h *handler) logout(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess.SetLanguage(user.Language)
	sess.SetTheme(user.Theme)

	if err := h.store.RemoveUserSessionByToken(user.ID, request.UserSessionToken(r)); err != nil {
		html.ServerError(w, r, err)
		return
	}

	http.SetCookie(w, cookie.Expired(
		cookie.CookieUserSessionID,
		config.Opts.HTTPS,
		config.Opts.BasePath(),
	))

	html.Redirect(w, r, route.Path(h.router, "login"))
}
