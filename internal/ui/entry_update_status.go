// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	json_parser "encoding/json"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) updateEntriesStatus(w http.ResponseWriter, r *http.Request) {
	var entriesStatusUpdateRequest model.EntriesStatusUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entriesStatusUpdateRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err := validator.ValidateEntriesStatusUpdateRequest(&entriesStatusUpdateRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	count, err := h.store.SetEntriesStatusCount(request.UserID(r), entriesStatusUpdateRequest.EntryIDs, entriesStatusUpdateRequest.Status)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, count)
}
