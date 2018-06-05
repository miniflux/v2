// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/http/response/xml"
	"github.com/miniflux/miniflux/reader/opml"
)

// Export generates the OPML file.
func (c *Controller) Export(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	opml, err := opml.NewHandler(c.store).Export(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	xml.Attachment(w, "feeds.opml", opml)
}
