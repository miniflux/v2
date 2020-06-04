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
	"miniflux.app/storage"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func showReadEntryPage(h *handler, w http.ResponseWriter, r *http.Request, readEntryOnly bool) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if entry == nil {
		html.NotFound(w, r)
		return
	}

	if entry.Status == model.EntryStatusUnread || entry.Status == model.EntryStatusMarked {
		err = h.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		entry.Status = model.EntryStatusRead
	}

	nextEntryRoute := ""
	prevEntryRoute := ""
	var prevEntry, nextEntry *model.Entry 
	entryPaginationBuilder := storage.NewEntryPaginationBuilder(h.store, user.ID, entry.ID, user.EntryDirection)
	if readEntryOnly {
		entryPaginationBuilder.WithStatus(model.EntryStatusRead)
		prevEntry, nextEntry, err = entryPaginationBuilder.Entries()
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		if nextEntry != nil {
			nextEntryRoute = route.Path(h.router, "historyReadEntry", "entryID", nextEntry.ID)
		}

		if prevEntry != nil {
			prevEntryRoute = route.Path(h.router, "historyReadEntry", "entryID", prevEntry.ID)
		}
	} else {
		entryPaginationBuilder.WithStatus(model.EntryStatusRead, model.EntryStatusMarked)
		prevEntry, nextEntry, err = entryPaginationBuilder.Entries()
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		if nextEntry != nil {
			nextEntryRoute = route.Path(h.router, "historyEntry", "entryID", nextEntry.ID)
		}

		if prevEntry != nil {
			prevEntryRoute = route.Path(h.router, "historyEntry", "entryID", prevEntry.ID)
		}
	}
	
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("entry", entry)
	view.Set("prevEntry", prevEntry)
	view.Set("nextEntry", nextEntry)
	view.Set("nextEntryRoute", nextEntryRoute)
	view.Set("prevEntryRoute", prevEntryRoute)
	view.Set("menu", "history")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("entry"))
}

func (h *handler) showHistoryEntryPage(w http.ResponseWriter, r *http.Request) {
	showReadEntryPage(h, w, r, /*readEntryOnly=*/false)
}

func (h *handler) showHistoryReadEntryPage(w http.ResponseWriter, r *http.Request) {
	showReadEntryPage(h, w, r, /*readEntryOnly=*/true)
}