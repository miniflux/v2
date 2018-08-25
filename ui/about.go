// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
	"miniflux.app/version"
)

// About shows the about page.
func (c *Controller) About(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)

	user, err := c.store.UserByID(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("version", version.Version)
	view.Set("build_date", version.BuildDate)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))

	html.OK(w, r, view.Render("about"))
}
