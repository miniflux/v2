// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/http/route"
)

// FlushHistory changes all "read" items to "removed".
func (c *Controller) FlushHistory(w http.ResponseWriter, r *http.Request) {
	err := c.store.FlushHistory(context.New(r).UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	response.Redirect(w, r, route.Path(c.router, "history"))
}
