// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
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
