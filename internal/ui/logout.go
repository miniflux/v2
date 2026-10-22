// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

func (h *handler) logout(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if s := request.WebSession(r); s != nil {
		if err := h.store.RemoveUserWebSession(user.ID, s.ID); err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
	}

	setPreferencesCookie(w, user.Language, user.Theme)

	response.HTMLRedirect(w, r, h.routePath("/"))
}
