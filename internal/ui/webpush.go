// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	json_parser "encoding/json"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
)

func (h *handler) registerWebpush(w http.ResponseWriter, r *http.Request) {
	var WebPushSubscriptionRequest model.WebPushSubscriptionRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&WebPushSubscriptionRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if err := h.store.RegisterWebPushSubscription(request.UserID(r), WebPushSubscriptionRequest); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSON(w, r, "OK")
}
