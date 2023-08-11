// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/ui/static"
)

func (h *handler) showStylesheet(w http.ResponseWriter, r *http.Request) {
	filename := request.RouteStringParam(r, "name")
	etag, found := static.StylesheetBundleChecksums[filename]
	if !found {
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(etag, 48*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Type", "text/css; charset=utf-8")
		b.WithBody(static.StylesheetBundles[filename])
		b.Write()
	})
}
