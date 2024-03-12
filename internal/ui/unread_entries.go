// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showUnreadPage(w http.ResponseWriter, r *http.Request) {
	beginPreProcessing := time.Now()
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	beginSqlCountUnreadEntries := time.Now()
	offset := request.QueryIntParam(r, "offset", 0)
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithGloballyVisible()
	countUnread, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	finishSqlCountUnreadEntries := time.Now()

	if offset >= countUnread {
		offset = 0
	}

	beginSqlFetchUnreadEntries := time.Now()
	builder = h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithSorting(user.EntryOrder, user.EntryDirection)
	builder.WithOffset(offset)
	builder.WithLimit(user.EntriesPerPage)
	builder.WithGloballyVisible()
	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	finishSqlFetchUnreadEntries := time.Now()

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("entries", entries)
	view.Set("pagination", getPagination(route.Path(h.router, "unread"), countUnread, offset, user.EntriesPerPage))
	view.Set("menu", "unread")
	view.Set("user", user)
	view.Set("countUnread", countUnread)
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	finishPreProcessing := time.Now()

	beginTemplateRendering := time.Now()
	render := view.Render("unread_entries")
	finishTemplateRendering := time.Now()

	if config.Opts.HasServerTimingHeader() {
		w.Header().Set("Server-Timing", fmt.Sprintf("pre_processing;dur=%d,sql_count_unread_entries;dur=%d,sql_fetch_unread_entries;dur=%d,template_rendering;dur=%d",
			finishPreProcessing.Sub(beginPreProcessing).Milliseconds(),
			finishSqlCountUnreadEntries.Sub(beginSqlCountUnreadEntries).Milliseconds(),
			finishSqlFetchUnreadEntries.Sub(beginSqlFetchUnreadEntries).Milliseconds(),
			finishTemplateRendering.Sub(beginTemplateRendering).Milliseconds(),
		))
	}

	html.OK(w, r, render)
}
