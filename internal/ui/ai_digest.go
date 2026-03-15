// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showAIDigestPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	offset := request.QueryIntParam(r, "offset", 0)

	// Count total unread entries with AI summaries for pagination.
	countBuilder := h.store.NewEntryQueryBuilder(user.ID)
	countBuilder.WithStatus(model.EntryStatusUnread)
	countBuilder.WithMinAIScore(1)
	total, err := countBuilder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if offset >= total {
		offset = 0
	}

	// Fetch paginated unread entries with AI summaries, sorted by score descending.
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithMinAIScore(1)
	builder.WithSorting("ai_score", "DESC")
	builder.WithSorting("id", "DESC")
	builder.WithOffset(offset)
	builder.WithLimit(user.EntriesPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("entries", entries)
	view.Set("total", total)
	view.Set("pagination", getPagination(route.Path(h.router, "unread"), total, offset, user.EntriesPerPage))
	view.Set("menu", "unread")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("showAIDigest", h.store.IsAIEnabled(user.ID))
	view.Set("countAIDigest", total)
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("ai_digest"))
}
