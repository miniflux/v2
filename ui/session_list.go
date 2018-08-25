// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/ui/session"
	"miniflux.app/ui/view"

	"miniflux.app/http/context"
	"miniflux.app/http/response/html"
)

// ShowSessions shows the list of active user sessions.
func (c *Controller) ShowSessions(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)

	user, err := c.store.UserByID(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	sessions, err := c.store.UserSessions(user.ID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	sessions.UseTimezone(user.Timezone)

	view.Set("currentSessionToken", ctx.UserSessionToken())
	view.Set("sessions", sessions)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))

	html.OK(w, r, view.Render("sessions"))
}
