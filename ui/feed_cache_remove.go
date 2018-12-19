// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
)

func (h *handler) removeFeedCaches(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if err := h.store.RemoveFeedCaches(request.UserID(r), feedID); err != nil {
		html.ServerError(w, r, err)
		return
	}
}
