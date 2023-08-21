// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/html"
)

func (h *handler) showIcon(w http.ResponseWriter, r *http.Request) {
	iconID := request.RouteInt64Param(r, "iconID")
	icon, err := h.store.IconByID(iconID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if icon == nil {
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(icon.Hash, 72*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Security-Policy", `default-src 'self'`)
		b.WithHeader("Content-Type", icon.MimeType)
		b.WithBody(icon.Content)
		b.WithoutCompression()
		b.Write()
	})
}
