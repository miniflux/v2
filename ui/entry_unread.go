// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/storage"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// ShowUnreadEntry shows a single feed entry in "unread" mode.
func (c *Controller) ShowUnreadEntry(w http.ResponseWriter, r *http.Request) {
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	entryID, err := request.IntParam(r, "entryID")
	if err != nil {
		html.BadRequest(w, err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(user.ID)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		html.ServerError(w, err)
		return
	}

	if entry == nil {
		html.NotFound(w)
		return
	}

	// Make sure we always get the pagination in unread mode even if the page is refreshed.
	if entry.Status == model.EntryStatusRead {
		err = c.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusUnread)
		if err != nil {
			logger.Error("[Controller:ShowUnreadEntry] %v", err)
			html.ServerError(w, nil)
			return
		}
	}

	entryPaginationBuilder := storage.NewEntryPaginationBuilder(c.store, user.ID, entry.ID, user.EntryDirection)
	entryPaginationBuilder.WithStatus(model.EntryStatusUnread)
	prevEntry, nextEntry, err := entryPaginationBuilder.Entries()
	if err != nil {
		html.ServerError(w, err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = route.Path(c.router, "unreadEntry", "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = route.Path(c.router, "unreadEntry", "entryID", prevEntry.ID)
	}

	// Always mark the entry as read after fetching the pagination.
	err = c.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
	if err != nil {
		logger.Error("[Controller:ShowUnreadEntry] %v", err)
		html.ServerError(w, nil)
		return
	}
	entry.Status = model.EntryStatusRead

	sess := session.New(c.store, request.SessionID(r))
	view := view.New(c.tpl, r, sess)
	view.Set("entry", entry)
	view.Set("prevEntry", prevEntry)
	view.Set("nextEntry", nextEntry)
	view.Set("nextEntryRoute", nextEntryRoute)
	view.Set("prevEntryRoute", prevEntryRoute)
	view.Set("menu", "unread")
	view.Set("user", user)
	view.Set("hasSaveEntry", c.store.HasSaveEntry(user.ID))
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))

	// Fetching the counter here avoid to be off by one.
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))

	html.OK(w, r, view.Render("entry"))
}
