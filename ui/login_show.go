// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/http/route"
	"github.com/miniflux/miniflux/ui/session"
	"github.com/miniflux/miniflux/ui/view"
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
	html.OK(w, view.Render("login"))
}
