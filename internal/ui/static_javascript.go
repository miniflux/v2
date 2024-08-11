// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/ui/static"
)

const licensePrefix = "//@license magnet:?xt=urn:btih:8e4f440f4c65981c5bf93c76d35135ba5064d8b7&dn=apache-2.0.txt Apache-2.0\n"
const licenseSuffix = "\n//@license-end"

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
			variables := fmt.Sprintf(`const OFFLINE_URL=%q;`, route.Path(h.router, "offline"))
			contents = append([]byte(variables), contents...)
		}

		// cloning the prefix since `append` mutates its first argument
		contents = append([]byte(strings.Clone(licensePrefix)), contents...)
		contents = append(contents, []byte(licenseSuffix)...)

		b.WithHeader("Content-Type", "text/javascript; charset=utf-8")
		b.WithBody(contents)
		b.Write()
	})
}
