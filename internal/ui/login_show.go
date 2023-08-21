// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showLoginPage(w http.ResponseWriter, r *http.Request) {
	if request.IsAuthenticated(r) {
		user, err := h.store.UserByID(request.UserID(r))
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		html.Redirect(w, r, route.Path(h.router, user.DefaultHomePage))
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	html.OK(w, r, view.Render("login"))
}
