// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/model"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// ShowUnreadPage render the page with all unread entries.
func (c *Controller) ShowUnreadPage(w http.ResponseWriter, r *http.Request) {
	sess := session.New(c.store, request.SessionID(r))
	view := view.New(c.tpl, r, sess)

	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	offset := request.QueryIntParam(r, "offset", 0)
	builder := c.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusUnread)
	countUnread, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, err)
		return
	}

	if offset >= countUnread {
		offset = 0
	}

	builder = c.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(user.EntryDirection)
	builder.WithOffset(offset)
	builder.WithLimit(nbItemsPerPage)
	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, err)
		return
	}

	view.Set("entries", entries)
	view.Set("pagination", c.getPagination(route.Path(c.router, "unread"), countUnread, offset))
	view.Set("menu", "unread")
	view.Set("user", user)
	view.Set("countUnread", countUnread)
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))
	view.Set("hasSaveEntry", c.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("unread_entries"))
}
