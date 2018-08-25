// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// ShowLoginPage shows the login form.
func (c *Controller) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	if ctx.IsAuthenticated() {
		response.Redirect(w, r, route.Path(c.router, "unread"))
		return
	}

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	html.OK(w, r, view.Render("login"))
}
