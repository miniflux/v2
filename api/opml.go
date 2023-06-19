// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/api"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/http/response/xml"
	"miniflux.app/reader/opml"
)

func (h *handler) exportFeeds(w http.ResponseWriter, r *http.Request) {
	opmlHandler := opml.NewHandler(h.store)
	opml, err := opmlHandler.Export(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	xml.OK(w, r, opml)
}

func (h *handler) importFeeds(w http.ResponseWriter, r *http.Request) {
	opmlHandler := opml.NewHandler(h.store)
	err := opmlHandler.Import(request.UserID(r), r.Body)
	defer r.Body.Close()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, map[string]string{"message": "Feeds imported successfully"})
}
