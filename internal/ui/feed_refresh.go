// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	feedHandler "miniflux.app/v2/internal/reader/handler"
)

func (h *handler) refreshFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	forceRefresh := request.QueryBoolParam(r, "forceRefresh", false)
	if err := feedHandler.RefreshFeed(h.store, request.UserID(r), feedID, forceRefresh); err != nil {
		slog.Warn("Unable to refresh feed",
			slog.Int64("user_id", request.UserID(r)),
			slog.Int64("feed_id", feedID),
			slog.Bool("force_refresh", forceRefresh),
			slog.Any("error", err),
		)
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
