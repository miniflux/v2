// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
)

func (h *handler) updateEntryVote(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	voteValue64 := request.RouteInt64Param(r, "vote")

	// Validate vote value
	if voteValue64 < -1 || voteValue64 > 1 {
		json.BadRequest(w, r, nil)
		return
	}

	voteValue := int(voteValue64)

	if err := h.store.UpdateEntryVote(request.UserID(r), entryID, voteValue); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, "OK")
}
