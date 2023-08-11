// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	json2 "encoding/json"
	"io"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
)

func (h *handler) saveEnclosureProgression(w http.ResponseWriter, r *http.Request) {
	enclosureID := request.RouteInt64Param(r, "enclosureID")
	enclosure, err := h.store.GetEnclosure(enclosureID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if enclosure == nil {
		json.NotFound(w, r)
		return
	}

	type enclosurePositionSaveRequest struct {
		Progression int64 `json:"progression"`
	}

	var postData enclosurePositionSaveRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json2.Unmarshal(body, &postData)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	enclosure.MediaProgression = postData.Progression

	err = h.store.UpdateEnclosure(enclosure)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, map[string]string{"message": "saved"})
}
