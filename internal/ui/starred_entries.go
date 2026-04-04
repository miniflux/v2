// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showStarredPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	offset := request.QueryIntParam(r, "offset", 0)
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithStarred(true)
	builder.WithSorting(user.EntryOrder, user.EntryDirection)
	builder.WithSorting("id", user.EntryDirection)
	builder.WithOffset(offset)
	builder.WithLimit(user.EntriesPerPage)

	entries, count, err := builder.GetEntriesWithCount()
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("total", count)
	view.Set("entries", entries)
	view.Set("pagination", getPagination(h.routePath("/starred"), count, offset, user.EntriesPerPage))
	view.Set("menu", "starred")
	view.Set("user", user)
	countUnread, countErrorFeeds, hasSaveEntry := h.store.GetNavMetadata(user.ID)
	view.Set("countUnread", countUnread)
	view.Set("countErrorFeeds", countErrorFeeds)
	view.Set("hasSaveEntry", hasSaveEntry)

	response.HTML(w, r, view.Render("starred_entries"))
}
