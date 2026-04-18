// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"path/filepath"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"

	"miniflux.app/v2/internal/ui/static"
)

func (h *handler) showAppIcon(w http.ResponseWriter, r *http.Request) {
	filename := request.RouteStringParam(r, "filename")
	bundle, ok := static.BinaryBundles[filename]
	if !ok {
		response.HTMLNotFound(w, r)
		return
	}

	response.NewBuilder(w, r).WithCaching(bundle.Checksum, 72*time.Hour, func(b *response.Builder) {
		body, encoding := bundle.Negotiate(r.Header.Get("Accept-Encoding"))
		b.WithoutCompression() // No need to compress already-compressed data.
		if encoding != "" {
			b.WithHeader("Content-Encoding", encoding)
			b.WithHeader("Vary", "Accept-Encoding")
		}
		switch filepath.Ext(filename) {
		case ".png":
			b.WithHeader("Content-Type", "image/png")
		case ".svg":
			b.WithHeader("Content-Type", "image/svg+xml")
		}
		b.WithBodyAsBytes(body)
		b.Write()
	})
}
