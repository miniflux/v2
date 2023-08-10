// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/ui"

import (
	"net/http"

	"miniflux.app/v2/http/request"
	"miniflux.app/v2/http/response/html"
	"miniflux.app/v2/http/response/xml"
	"miniflux.app/v2/reader/opml"
)

func (h *handler) exportFeeds(w http.ResponseWriter, r *http.Request) {
	opml, err := opml.NewHandler(h.store).Export(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	xml.Attachment(w, r, "feeds.opml", opml)
}
