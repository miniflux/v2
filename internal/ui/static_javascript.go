// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/ui/static"
)

func (h *handler) showJavascript(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	bundle, found := static.JavascriptBundles[filename]
	if !found {
		response.HTMLNotFound(w, r)
		return
	}

	response.NewBuilder(w, r).WithCaching(bundle.Checksum, 48*time.Hour, func(b *response.Builder) {
		body, encoding := bundle.Negotiate(r.Header.Get("Accept-Encoding"))
		b.WithoutCompression()
		b.WithHeader("Content-Type", "text/javascript; charset=utf-8")
		b.WithHeader("Vary", "Accept-Encoding")
		if encoding != "" {
			b.WithHeader("Content-Encoding", encoding)
		}
		b.WithBodyAsBytes(body)
		b.Write()
	})
}
