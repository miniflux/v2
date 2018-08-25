// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
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
