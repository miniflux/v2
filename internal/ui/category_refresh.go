// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
)

func (h *handler) refreshCategoryEntriesPage(w http.ResponseWriter, r *http.Request) {
	categoryID := h.refreshCategory(w, r)
	html.Redirect(w, r, route.Path(h.router, "categoryEntries", "categoryID", categoryID))
}

func (h *handler) refreshCategoryFeedsPage(w http.ResponseWriter, r *http.Request) {
	categoryID := h.refreshCategory(w, r)
	html.Redirect(w, r, route.Path(h.router, "categoryFeeds", "categoryID", categoryID))
}

func (h *handler) refreshCategory(w http.ResponseWriter, r *http.Request) int64 {
	userID := request.UserID(r)
	categoryID := request.RouteInt64Param(r, "categoryID")

	jobs, err := h.store.NewCategoryBatch(userID, categoryID, h.store.CountFeeds(userID))
	if err != nil {
		html.ServerError(w, r, err)
		return 0
	}

	slog.Info(
		"Triggered a manual refresh of all feeds for a given category from the web ui",
		slog.Int64("user_id", userID),
		slog.Int64("category_id", categoryID),
		slog.Int("nb_jobs", len(jobs)),
	)

	go h.pool.Push(jobs)

	return categoryID
}
