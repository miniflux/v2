// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/logger"
)

func (h *handler) removeSession(w http.ResponseWriter, r *http.Request) {
	sessionID := request.RouteInt64Param(r, "sessionID")
	err := h.store.RemoveUserSessionByID(request.UserID(r), sessionID)
	if err != nil {
		logger.Error("[UI:RemoveSession] %v", err)
	}

	html.Redirect(w, r, route.Path(h.router, "sessions"))
}
