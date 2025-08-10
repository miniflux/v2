// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "influxeed-engine/v2/internal/ui"

import (
	"net/http"

	"influxeed-engine/v2/internal/http/request"
	"influxeed-engine/v2/internal/http/response/html"
	"influxeed-engine/v2/internal/ui/session"
	"influxeed-engine/v2/internal/ui/view"
)

func (h *handler) showOfflinePage(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	html.OK(w, r, view.Render("offline"))
}
