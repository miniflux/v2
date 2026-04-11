// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showLoginPage(w http.ResponseWriter, r *http.Request) {
	if request.IsAuthenticated(r) {
		user, err := h.store.UserByID(request.UserID(r))
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		response.HTMLRedirect(w, r, h.basePath+"/"+user.DefaultHomePage)
		return
	}

	view := view.New(h.tpl, r)
	redirectURL := request.QueryStringParam(r, "redirect_url", "")
	view.Set("redirectURL", redirectURL)
	response.HTML(w, r, view.Render("login"))
}
