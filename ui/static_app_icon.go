// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"
	"path/filepath"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/static"
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
