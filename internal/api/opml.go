// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/http/response/xml"
	"miniflux.app/v2/internal/reader/opml"
)

func (h *handler) exportFeeds(w http.ResponseWriter, r *http.Request) {
	opmlHandler := opml.NewHandler(h.store)
	opmlExport, err := opmlHandler.Export(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	xml.OK(w, r, opmlExport)
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
