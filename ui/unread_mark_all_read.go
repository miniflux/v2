// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
)

// MarkAllAsRead marks all unread entries as read.
func (c *Controller) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	if err := c.store.MarkAllAsRead(request.UserID(r)); err != nil {
		logger.Error("[MarkAllAsRead] %v", err)
	}

	html.Redirect(w, r, route.Path(c.router, "unread"))
}
