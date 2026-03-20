// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/http/response"

	"miniflux.app/v2/internal/ui/static"
)

func (h *handler) showStylesheet(w http.ResponseWriter, r *http.Request) {
	// The filename path value contains "name.checksum.css"; extract the name portion.
	filename, _, _ := strings.Cut(r.PathValue("filename"), ".")
	m, found := static.StylesheetBundles[filename]
	if !found {
		response.HTMLNotFound(w, r)
		return
	}

	response.NewBuilder(w, r).WithCaching(m.Checksum, 48*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Type", "text/css; charset=utf-8")
		b.WithBodyAsBytes(m.Data)
		b.Write()
	})
}
