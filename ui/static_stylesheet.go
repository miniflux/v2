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

// Stylesheet renders the CSS.
func (c *Controller) Stylesheet(w http.ResponseWriter, r *http.Request) {
	stylesheet := request.RouteStringParam(r, "name")
	if _, found := static.Stylesheets[stylesheet]; !found {
		html.NotFound(w)
		return
	}

	body := static.Stylesheets[stylesheet]
	etag := static.StylesheetsChecksums[stylesheet]

	response.Cache(w, r, "text/css; charset=utf-8", etag, []byte(body), 48*time.Hour)
}
