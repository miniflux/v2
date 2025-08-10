// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "influxeed-engine/v2/internal/ui"

import (
	"net/http"

	"influxeed-engine/v2/internal/http/request"
	"influxeed-engine/v2/internal/http/response/json"
)

func (h *handler) markAllAsRead(w http.ResponseWriter, r *http.Request) {
	if err := h.store.MarkGloballyVisibleFeedsAsRead(request.UserID(r)); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, "OK")
}
