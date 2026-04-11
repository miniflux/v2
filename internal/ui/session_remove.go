// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

func (h *handler) removeSession(w http.ResponseWriter, r *http.Request) {
	sessionID := request.RouteStringParam(r, "sessionID")
	err := h.store.RemoveUserWebSession(request.UserID(r), sessionID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/sessions"))
}
