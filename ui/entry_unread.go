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

func (h *handler) showUnreadEntryPage(w http.ResponseWriter, r *http.Request) {
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
		html.Redirect(w, r, route.Path(h.router, "unread"))
		return
	}

	// Make sure we always get the pagination in unread mode even if the page is refreshed.
	if entry.Status == model.EntryStatusRead {
		err = h.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusUnread)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	}

	entryPaginationBuilder := storage.NewEntryPaginationBuilder(h.store, user.ID, entry.ID, user.EntryOrder, user.EntryDirection)
	entryPaginationBuilder.WithStatus(model.EntryStatusUnread)
	entryPaginationBuilder.WithGloballyVisible()
	prevEntry, nextEntry, err := entryPaginationBuilder.Entries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = route.Path(h.router, "unreadEntry", "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = route.Path(h.router, "unreadEntry", "entryID", prevEntry.ID)
	}

	// Always mark the entry as read after fetching the pagination.
	err = h.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	entry.Status = model.EntryStatusRead

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("entry", entry)
	view.Set("prevEntry", prevEntry)
	view.Set("nextEntry", nextEntry)
	view.Set("nextEntryRoute", nextEntryRoute)
	view.Set("prevEntryRoute", prevEntryRoute)
	view.Set("menu", "unread")
	view.Set("user", user)
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	// Fetching the counter here avoid to be off by one.
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))

	html.OK(w, r, view.Render("entry"))
}
