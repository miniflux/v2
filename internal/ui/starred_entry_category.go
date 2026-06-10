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

func (h *handler) showStarredCategoryEntryPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	categoryID := request.RouteInt64Param(r, "categoryID")
	entryID := request.RouteInt64Param(r, "entryID")

	entry, err := h.store.NewEntryQueryBuilder(user.ID).
		WithCategoryID(categoryID).
		WithEntryIDs(entryID).
		GetEntry()
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if entry == nil {
		response.HTMLNotFound(w, r)
		return
	}

	if entry.ShouldMarkAsReadOnView(user) {
		err = h.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		entry.Status = model.EntryStatusRead
	}

	if user.AlwaysOpenExternalLinks {
		response.HTMLRedirect(w, r, entry.URL)
		return
	}

	prevEntry, nextEntry, err := h.store.NewEntryPaginationBuilder(user.ID, entry.ID, user.EntryOrder, user.EntryDirection).
		WithCategoryID(categoryID).
		WithStarred().
		Entries()
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = h.routePath("/starred/category/%d/entry/%d", categoryID, nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = h.routePath("/starred/category/%d/entry/%d", categoryID, prevEntry.ID)
	}

	view := view.New(h.tpl, r)
	view.Set("entry", entry)
	view.Set("prevEntry", prevEntry)
	view.Set("nextEntry", nextEntry)
	view.Set("nextEntryRoute", nextEntryRoute)
	view.Set("prevEntryRoute", prevEntryRoute)
	view.Set("menu", "categories")
	view.Set("user", user)
	navMetadata, _ := h.store.GetNavMetadata(user.ID)
	view.Set("countUnread", navMetadata.CountUnread)
	view.Set("countErrorFeeds", navMetadata.CountErrorFeeds)
	view.Set("hasSaveEntry", navMetadata.HasSaveEntry)

	response.HTML(w, r, view.Render("entry"))
}
