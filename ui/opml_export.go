// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/response/xml"
	"miniflux.app/reader/opml"
)

// Export generates the OPML file.
func (c *Controller) Export(w http.ResponseWriter, r *http.Request) {
	opml, err := opml.NewHandler(c.store).Export(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	xml.Attachment(w, r, "feeds.opml", opml)
}
