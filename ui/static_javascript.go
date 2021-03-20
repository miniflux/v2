// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"fmt"
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/ui/static"
)

func (h *handler) showJavascript(w http.ResponseWriter, r *http.Request) {
	filename := request.RouteStringParam(r, "name")
	etag, found := static.JavascriptBundleChecksums[filename]
	if !found {
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(etag, 48*time.Hour, func(b *response.Builder) {
		contents := static.JavascriptBundles[filename]

		if filename == "service-worker" {
			variables := fmt.Sprintf(`const OFFLINE_URL="%s";`, route.Path(h.router, "offline"))
			contents = append([]byte(variables)[:], contents[:]...)
		}

		b.WithHeader("Content-Type", "text/javascript; charset=utf-8")
		b.WithBody(contents)
		b.Write()
	})
}
