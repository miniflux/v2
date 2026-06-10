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

func (h *handler) showUnreadPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	offset := request.QueryIntParam(r, "offset", 0)

	entries, countUnread, err := h.store.NewEntryQueryBuilder(user.ID).
		WithStatuses(model.EntryStatusUnread).
		WithSorting(user.EntryOrder, user.EntryDirection).
		WithSorting("id", user.EntryDirection).
		WithOffset(offset).
		WithLimit(user.EntriesPerPage).
		WithGloballyVisible().
		WithoutContent().
		GetEntriesWithCount()
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if offset >= countUnread && countUnread > 0 {
		offset = 0

		entries, countUnread, err = h.store.NewEntryQueryBuilder(user.ID).
			WithStatuses(model.EntryStatusUnread).
			WithSorting(user.EntryOrder, user.EntryDirection).
			WithSorting("id", user.EntryDirection).
			WithLimit(user.EntriesPerPage).
			WithGloballyVisible().
			WithoutContent().
			GetEntriesWithCount()
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
	}

	view := view.New(h.tpl, r)
	view.Set("entries", entries)
	view.Set("pagination", getPagination(h.routePath("/unread"), countUnread, offset, user.EntriesPerPage))
	view.Set("menu", "unread")
	view.Set("user", user)
	navMetadata, _ := h.store.GetNavMetadata(user.ID)
	view.Set("countUnread", countUnread)
	view.Set("countErrorFeeds", navMetadata.CountErrorFeeds)
	view.Set("hasSaveEntry", navMetadata.HasSaveEntry)

	response.HTML(w, r, view.Render("unread_entries"))
}
