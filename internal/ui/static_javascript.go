// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/http/response"

	"miniflux.app/v2/internal/ui/static"
)

const licensePrefix = "//@license magnet:?xt=urn:btih:8e4f440f4c65981c5bf93c76d35135ba5064d8b7&dn=apache-2.0.txt Apache-2.0\n"
const licenseSuffix = "\n//@license-end"

func (h *handler) showJavascript(w http.ResponseWriter, r *http.Request) {
	// The filename path value contains "name.checksum.js"; reject non-JS requests
	// and extract the name portion.
	rawFilename := r.PathValue("filename")
	if !strings.HasSuffix(rawFilename, ".js") {
		response.HTMLNotFound(w, r)
		return
	}
	filename, _, _ := strings.Cut(strings.TrimSuffix(rawFilename, ".js"), ".")
	js, found := static.JavascriptBundles[filename]
	if !found {
		response.HTMLNotFound(w, r)
		return
	}

	response.NewBuilder(w, r).WithCaching(js.Checksum, 48*time.Hour, func(b *response.Builder) {
		contents := js.Data

		if filename == "service-worker" {
			variables := fmt.Sprintf(`const OFFLINE_URL=%q;`, h.routePath("/offline"))
			contents = append([]byte(variables), contents...)
		}

		// cloning the prefix since `append` mutates its first argument
		contents = append([]byte(strings.Clone(licensePrefix)), contents...)
		contents = append(contents, []byte(licenseSuffix)...)

		b.WithHeader("Content-Type", "text/javascript; charset=utf-8")
		b.WithBodyAsBytes(contents)
		b.Write()
	})
}
