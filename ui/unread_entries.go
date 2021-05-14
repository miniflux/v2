// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/model"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"

	servertiming "github.com/mitchellh/go-server-timing"
)

func (h *handler) showUnreadPage(w http.ResponseWriter, r *http.Request) {
	timing := servertiming.FromContext(r.Context())
	prep := timing.NewMetric("pre_processing").Start()

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	m := timing.NewMetric("sql_count_unread_entries").Start()
	offset := request.QueryIntParam(r, "offset", 0)
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithGloballyVisible()
	countUnread, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	m.Stop()

	if offset >= countUnread {
		offset = 0
	}

	m = timing.NewMetric("sql_fetch_unread_entries").Start()
	builder = h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithOrder(user.EntryOrder)
	builder.WithDirection(user.EntryDirection)
	builder.WithOffset(offset)
	builder.WithLimit(user.EntriesPerPage)
	builder.WithGloballyVisible()
	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	m.Stop()

	view.Set("entries", entries)
	view.Set("pagination", getPagination(route.Path(h.router, "unread"), countUnread, offset, user.EntriesPerPage))
	view.Set("menu", "unread")
	view.Set("user", user)
	view.Set("countUnread", countUnread)
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	prep.Stop()

	m = timing.NewMetric("template_rendering").Start()
	render := view.Render("unread_entries")
	m.Stop()

	html.OK(w, r, render)
}
