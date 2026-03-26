// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/reader/opml"
)

func (h *handler) exportFeedsHandler(w http.ResponseWriter, r *http.Request) {
	opmlHandler := opml.NewHandler(h.store)
	opmlExport, err := opmlHandler.Export(request.UserID(r))
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.XML(w, r, opmlExport)
}

func (h *handler) importFeedsHandler(w http.ResponseWriter, r *http.Request) {
	opmlHandler := opml.NewHandler(h.store)
	err := opmlHandler.Import(request.UserID(r), r.Body)
	defer r.Body.Close()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSONCreated(w, r, importFeedsResponse{Message: "Feeds imported successfully"})
}
