// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/response/html"
	"miniflux.app/ui/session"
	"miniflux.app/http/request"
	"miniflux.app/ui/view"
)

// CategoryList shows the page with all categories.
func (c *Controller) CategoryList(w http.ResponseWriter, r *http.Request) {
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	categories, err := c.store.CategoriesWithFeedCount(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(c.store, request.SessionID(r))
	view := view.New(c.tpl, r, sess)
	view.Set("categories", categories)
	view.Set("total", len(categories))
	view.Set("menu", "categories")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))

	html.OK(w, r, view.Render("categories"))
}
