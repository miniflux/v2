// Copyright 2017 Frédéric Guillot. All rights reserved.
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
)

func showHistory(h *handler, w http.ResponseWriter, r *http.Request, readEntriesOnly bool) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	offset := request.QueryIntParam(r, "offset", 0)
	builder := h.store.NewEntryQueryBuilder(user.ID)
	if readEntriesOnly {
		builder.WithStatus(model.EntryStatusRead)
	} else {
		builder.WithStatus(model.EntryStatusRead, model.EntryStatusMarked)
	}
	builder.WithOrder("changed_at")
	builder.WithDirection("desc")
	builder.WithOffset(offset)
	builder.WithLimit(nbItemsPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("entries", entries)
	view.Set("display_count", count)
	if readEntriesOnly {
		view.Set("pagination", getPagination(route.Path(h.router, "historyReadEntries"), count, offset))
	} else {
		view.Set("pagination", getPagination(route.Path(h.router, "history"), count, offset))
	}
	view.Set("menu", "history")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))
	view.Set("showOnlyReadEntries", readEntriesOnly)

	html.OK(w, r, view.Render("history_entries"))
}

func (h *handler) showHistoryPage(w http.ResponseWriter, r *http.Request) {
	showHistory(h, w, r, /*readEntriesOnly=*/false)
}

func (h *handler) showHistoryReadEntriesPage(w http.ResponseWriter, r *http.Request) {
	showHistory(h, w, r, /*readEntriesOnly=*/true)
}
