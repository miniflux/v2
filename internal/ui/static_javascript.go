// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/ui/static"
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
			user, err := h.store.UserByID(request.UserID(r))
			if err != nil {
				html.ServerError(w, r, err)
				return
			}

			cacheForOffline := 0
			if user.CacheForOffline {
				cacheForOffline = 1
			}

			variables := fmt.Sprintf(`const OFFLINE_URL=%q;const USE_CACHE=%d;`, route.Path(h.router, "offline"), cacheForOffline)

			contents = append([]byte(variables), contents...)
		}

		b.WithHeader("Content-Type", "text/javascript; charset=utf-8")
		b.WithBody(contents)
		b.Write()
	})
}
