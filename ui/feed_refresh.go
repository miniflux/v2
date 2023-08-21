// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	feedHandler "miniflux.app/reader/handler"
)

func (h *handler) refreshFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if err := feedHandler.RefreshFeed(h.store, request.UserID(r), feedID); err != nil {
		logger.Error("[UI:RefreshFeed] %v", err)
	}

	html.Redirect(w, r, route.Path(h.router, "feedEntries", "feedID", feedID))
}

func (h *handler) refreshAllFeeds(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	jobs, err := h.store.NewUserBatch(userID, h.store.CountFeeds(userID))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	go func() {
		h.pool.Push(jobs)
	}()

	html.Redirect(w, r, route.Path(h.router, "feeds"))
}
