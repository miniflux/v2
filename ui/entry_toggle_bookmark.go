// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/logger"
)

// ToggleBookmark handles Ajax request to toggle bookmark value.
func (c *Controller) ToggleBookmark(w http.ResponseWriter, r *http.Request) {
	entryID, err := request.IntParam(r, "entryID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	if err := c.store.ToggleBookmark(request.UserID(r), entryID); err != nil {
		logger.Error("[Controller:ToggleBookmark] %v", err)
		json.ServerError(w, nil)
		return
	}

	json.OK(w, r, "OK")
}
