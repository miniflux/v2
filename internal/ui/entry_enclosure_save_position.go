// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	json_parser "encoding/json"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

func (h *handler) saveEnclosureProgression(w http.ResponseWriter, r *http.Request) {
	enclosureID := request.RouteInt64Param(r, "enclosureID")
	enclosure, err := h.store.GetEnclosure(enclosureID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if enclosure == nil {
		response.JSONNotFound(w, r)
		return
	}

	type enclosurePositionSaveRequest struct {
		Progression int64 `json:"progression"`
	}

	var postData enclosurePositionSaveRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&postData); err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	enclosure.MediaProgression = postData.Progression

	if err := h.store.UpdateEnclosure(enclosure); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSONCreated(w, r, map[string]string{"message": "saved"})
}
