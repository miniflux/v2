// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/response/json"
	"github.com/miniflux/miniflux/http/response/xml"
	"github.com/miniflux/miniflux/reader/opml"
)

// Export is the API handler that export feeds to OPML.
func (c *Controller) Export(w http.ResponseWriter, r *http.Request) {
	opmlHandler := opml.NewHandler(c.store)
	opml, err := opmlHandler.Export(context.New(r).UserID())
	if err != nil {
		json.ServerError(w, err)
		return
	}

	xml.OK(w, opml)
}

// Import is the API handler that import an OPML file.
func (c *Controller) Import(w http.ResponseWriter, r *http.Request) {
	opmlHandler := opml.NewHandler(c.store)
	err := opmlHandler.Import(context.New(r).UserID(), r.Body)
	defer r.Body.Close()
	if err != nil {
		json.ServerError(w, err)
		return
	}

	json.Created(w, map[string]string{"message": "Feeds imported successfully"})
}
