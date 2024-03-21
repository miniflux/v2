// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/mediaproxy"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/processor"
	"miniflux.app/v2/internal/storage"
)

func (h *handler) fetchContent(w http.ResponseWriter, r *http.Request) {
	loggedUserID := request.UserID(r)
	entryID := request.RouteInt64Param(r, "entryID")

	entryBuilder := h.store.NewEntryQueryBuilder(loggedUserID)
	entryBuilder.WithEntryID(entryID)
	entryBuilder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := entryBuilder.GetEntry()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if entry == nil {
		json.NotFound(w, r)
		return
	}

	user, err := h.store.UserByID(loggedUserID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	feedBuilder := storage.NewFeedQueryBuilder(h.store, loggedUserID)
	feedBuilder.WithFeedID(entry.FeedID)
	feed, err := feedBuilder.GetFeed()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if feed == nil {
		json.NotFound(w, r)
		return
	}

	if err := processor.ProcessEntryWebPage(feed, entry, user); err != nil {
		json.ServerError(w, r, err)
		return
	}

	if err := h.store.UpdateEntryTitleAndContent(entry); err != nil {
		json.ServerError(w, r, err)
		return
	}

	readingTime := locale.NewPrinter(user.Language).Plural("entry.estimated_reading_time", entry.ReadingTime, entry.ReadingTime)

	json.OK(w, r, map[string]string{"content": mediaproxy.RewriteDocumentWithRelativeProxyURL(h.router, entry.Content), "reading_time": readingTime})
}
