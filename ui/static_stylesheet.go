// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/static"
)

func (h *handler) showStylesheet(w http.ResponseWriter, r *http.Request) {
	filename := request.RouteStringParam(r, "name")
	if filename == "custom_css" {
		user, err := h.store.UserByID(request.UserID(r))
		if err != nil {
			html.NotFound(w, r)
			return
		}
		b := response.New(w, r)
		if user == nil {
			b.WithHeader("Content-Type", "text/css; charset=utf-8")
			b.WithBody("")
			b.Write()
			return
		}
		b.WithHeader("Content-Type", "text/css; charset=utf-8")
		b.WithBody(user.Extra["custom_css"])
		b.Write()
		return
	}
	etag, found := static.StylesheetsChecksums[filename]
	if !found {
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(etag, 48*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Type", "text/css; charset=utf-8")
		b.WithBody(static.Stylesheets[filename])
		b.Write()
	})
}
