// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
)

func (h *handler) markAllAsRead(w http.ResponseWriter, r *http.Request) {
	if err := h.store.MarkAllAsRead(request.UserID(r)); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, "OK")
}
