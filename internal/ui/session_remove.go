// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "influxeed-engine/v2/internal/ui"

import (
	"net/http"

	"influxeed-engine/v2/internal/http/request"
	"influxeed-engine/v2/internal/http/response/html"
	"influxeed-engine/v2/internal/http/route"
)

func (h *handler) removeSession(w http.ResponseWriter, r *http.Request) {
	sessionID := request.RouteInt64Param(r, "sessionID")
	err := h.store.RemoveUserSessionByID(request.UserID(r), sessionID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "sessions"))
}
