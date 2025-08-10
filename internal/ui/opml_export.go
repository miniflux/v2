// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "influxeed-engine/v2/internal/ui"

import (
	"net/http"

	"influxeed-engine/v2/internal/http/request"
	"influxeed-engine/v2/internal/http/response/html"
	"influxeed-engine/v2/internal/http/response/xml"
	"influxeed-engine/v2/internal/reader/opml"
)

func (h *handler) exportFeeds(w http.ResponseWriter, r *http.Request) {
	opmlExport, err := opml.NewHandler(h.store).Export(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	xml.Attachment(w, r, "feeds.opml", opmlExport)
}
