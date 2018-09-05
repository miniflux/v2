// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/static"
)

// Javascript renders application client side code.
func (c *Controller) Javascript(w http.ResponseWriter, r *http.Request) {
	filename := request.Param(r, "name", "app")
	if _, found := static.Javascripts[filename]; !found {
		html.NotFound(w)
		return
	}

	body := static.Javascripts[filename]
	etag := static.JavascriptsChecksums[filename]

	response.Cache(w, r, "text/javascript; charset=utf-8", etag, []byte(body), 48*time.Hour)
}
