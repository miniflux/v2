// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/locale"
)

func (h *handler) refreshCategoryEntriesPage(w http.ResponseWriter, r *http.Request) {
	categoryID := h.refreshCategory(w, r)
	response.HTMLRedirect(w, r, h.routePath("/category/%d/entries", categoryID))
}

func (h *handler) refreshCategoryFeedsPage(w http.ResponseWriter, r *http.Request) {
	categoryID := h.refreshCategory(w, r)
	response.HTMLRedirect(w, r, h.routePath("/category/%d/feeds", categoryID))
}

func (h *handler) refreshCategory(w http.ResponseWriter, r *http.Request) int64 {
	categoryID := request.RouteInt64Param(r, "categoryID")
	sess := request.WebSession(r)
	printer := locale.NewPrinter(sess.Language())

	// Avoid accidental and excessive refreshes.
	if time.Since(sess.LastForceRefresh()) < config.Opts.ForceRefreshInterval() {
		interval := int(config.Opts.ForceRefreshInterval().Minutes())
		sess.SetErrorMessage(printer.Plural("alert.too_many_feeds_refresh", interval, interval))
	} else {
		userID := request.UserID(r)
		// We allow the end-user to force refresh all its feeds in this category
		// without taking into consideration the number of errors.
		batchBuilder := h.store.NewBatchBuilder()
		batchBuilder.WithoutDisabledFeeds()
		batchBuilder.WithUserID(userID)
		batchBuilder.WithCategoryID(categoryID)
		batchBuilder.WithLimitPerHost(config.Opts.PollingLimitPerHost())

		jobs, err := batchBuilder.FetchJobs()
		if err != nil {
			response.HTMLServerError(w, r, err)
			return 0
		}

		slog.Info(
			"Triggered a manual refresh of all feeds for a given category from the web ui",
			slog.Int64("user_id", userID),
			slog.Int64("category_id", categoryID),
			slog.Int("nb_jobs", len(jobs)),
		)

		go h.pool.Push(jobs)

		sess.MarkForceRefreshed()
		sess.SetSuccessMessage(printer.Print("alert.background_feed_refresh"))
	}

	return categoryID
}
