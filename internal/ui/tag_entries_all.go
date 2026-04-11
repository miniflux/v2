// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"net/url"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showTagEntriesAllPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	tagName, err := url.PathUnescape(request.RouteStringParam(r, "tagName"))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	offset := request.QueryIntParam(r, "offset", 0)
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithTags([]string{tagName})
	builder.WithSorting("status", "asc")
	builder.WithSorting(user.EntryOrder, user.EntryDirection)
	builder.WithSorting("id", user.EntryDirection)
	builder.WithoutContent()
	builder.WithOffset(offset)
	builder.WithLimit(user.EntriesPerPage)

	entries, count, err := builder.GetEntriesWithCount()
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	view := view.New(h.tpl, r)
	view.Set("tagName", tagName)
	view.Set("total", count)
	view.Set("entries", entries)
	view.Set("pagination", getPagination(h.routePath("/tags/%s/entries/all", url.PathEscape(tagName)), count, offset, user.EntriesPerPage))
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))
	view.Set("showOnlyUnreadEntries", false)

	response.HTML(w, r, view.Render("tag_entries"))
}
