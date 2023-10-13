// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/response/xml"
	"miniflux.app/v2/internal/reader/opml"
)

func (h *handler) exportFeeds(w http.ResponseWriter, r *http.Request) {
	opmlExport, err := opml.NewHandler(h.store).Export(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	xml.Attachment(w, r, "feeds.opml", opmlExport)
}
