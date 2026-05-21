// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showUnreadEntryPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")

	entry, err := h.store.NewEntryQueryBuilder(user.ID).
		WithEntryIDs(entryID).
		GetEntry()
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if entry == nil {
		response.HTMLRedirect(w, r, h.routePath("/unread"))
		return
	}

	// Make sure we always get the pagination in unread mode even if the page is refreshed.
	if entry.Status == model.EntryStatusRead {
		err = h.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusUnread)
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
	}

	prevEntry, nextEntry, err := h.store.NewEntryPaginationBuilder(user.ID, entry.ID, user.EntryOrder, user.EntryDirection).
		WithStatus(model.EntryStatusUnread).
		WithGloballyVisible().
		Entries()
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = h.routePath("/unread/entry/%d", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = h.routePath("/unread/entry/%d", prevEntry.ID)
	}

	if entry.ShouldMarkAsReadOnView(user) {
		entry.Status = model.EntryStatusRead
	}

	// Restore entry read status if needed after fetching the pagination.
	if entry.Status == model.EntryStatusRead {
		err = h.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
	}

	if user.AlwaysOpenExternalLinks {
		response.HTMLRedirect(w, r, entry.URL)
		return
	}

	view := view.New(h.tpl, r)
	view.Set("entry", entry)
	view.Set("prevEntry", prevEntry)
	view.Set("nextEntry", nextEntry)
	view.Set("nextEntryRoute", nextEntryRoute)
	view.Set("prevEntryRoute", prevEntryRoute)
	view.Set("menu", "unread")
	view.Set("user", user)
	// Fetching the counters here avoids being off by one.
	navMetadata, _ := h.store.GetNavMetadata(user.ID)
	view.Set("countUnread", navMetadata.CountUnread)
	view.Set("countErrorFeeds", navMetadata.CountErrorFeeds)
	view.Set("hasSaveEntry", navMetadata.HasSaveEntry)

	response.HTML(w, r, view.Render("entry"))
}
