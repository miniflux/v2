// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/ui/static"
)

func (h *handler) showFavicon(w http.ResponseWriter, r *http.Request) {
	value, ok := static.BinaryBundles["favicon.ico"]
	if !ok {
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(value.Checksum, 48*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Type", "image/x-icon")
		b.WithoutCompression()
		b.WithBody(value.Data)
		b.Write()
	})
}
