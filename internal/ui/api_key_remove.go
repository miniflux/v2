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

func (h *handler) removeAPIKey(w http.ResponseWriter, r *http.Request) {
	keyID := request.RouteInt64Param(r, "keyID")
	err := h.store.RemoveAPIKey(request.UserID(r), keyID)
	if err != nil {
		logger.Error("[UI:RemoveAPIKey] %v", err)
	}

	html.Redirect(w, r, route.Path(h.router, "apiKeys"))
}
