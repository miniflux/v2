// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"path/filepath"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/ui/static"
)

func (h *handler) showAppIcon(w http.ResponseWriter, r *http.Request) {
	filename := request.RouteStringParam(r, "filename")
	etag, err := static.GetBinaryFileChecksum(filename)
	if err != nil {
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
		blob, err := static.LoadBinaryFile(filename)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		switch filepath.Ext(filename) {
		case ".png":
			b.WithHeader("Content-Type", "image/png")
		case ".svg":
			b.WithHeader("Content-Type", "image/svg+xml")
		}

		b.WithoutCompression()
		b.WithBody(blob)
		b.Write()
	})
}
