// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"encoding/base64"
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/static"
)

// AppIcon shows application icons.
func (c *Controller) AppIcon(w http.ResponseWriter, r *http.Request) {
	filename := request.RouteStringParam(r, "filename")
	etag, found := static.BinariesChecksums[filename]
	if !found {
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
		blob, err := base64.StdEncoding.DecodeString(static.Binaries[filename])
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		b.WithHeader("Content-Type", "image/png")
		b.WithoutCompression()
		b.WithBody(blob)
		b.Write()
	})
}
