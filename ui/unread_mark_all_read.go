// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/route"
	"github.com/miniflux/miniflux/logger"
)

// MarkAllAsRead marks all unread entries as read.
func (c *Controller) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	if err := c.store.MarkAllAsRead(context.New(r).UserID()); err != nil {
		logger.Error("[MarkAllAsRead] %v", err)
	}

	response.Redirect(w, r, route.Path(c.router, "unread"))
}
