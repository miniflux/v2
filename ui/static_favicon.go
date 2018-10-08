// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"encoding/base64"
	"net/http"
	"time"

	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/static"
)

// Favicon shows the application favicon.
func (c *Controller) Favicon(w http.ResponseWriter, r *http.Request) {
	etag, found := static.BinariesChecksums["favicon.ico"]
	if !found {
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(etag, 48*time.Hour, func(b *response.Builder) {
		blob, err := base64.StdEncoding.DecodeString(static.Binaries["favicon.ico"])
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		b.WithHeader("Content-Type", "image/x-icon")
		b.WithoutCompression()
		b.WithBody(blob)
		b.Write()
	})
}
