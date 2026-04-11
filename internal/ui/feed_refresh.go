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
	feedHandler "miniflux.app/v2/internal/reader/handler"
)

func (h *handler) refreshFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	forceRefresh := request.QueryBoolParam(r, "forceRefresh", false)
	if localizedError := feedHandler.RefreshFeed(h.store, request.UserID(r), feedID, forceRefresh); localizedError != nil {
		slog.Warn("Unable to refresh feed",
			slog.Int64("user_id", request.UserID(r)),
			slog.Int64("feed_id", feedID),
			slog.Bool("force_refresh", forceRefresh),
			slog.Any("error", localizedError.Error()),
		)
	}

	response.HTMLRedirect(w, r, h.routePath("/feed/%d/entries", feedID))
}

func (h *handler) refreshAllFeeds(w http.ResponseWriter, r *http.Request) {
	sess := request.WebSession(r)
	printer := locale.NewPrinter(sess.Language())

	// Avoid accidental and excessive refreshes.
	if time.Since(sess.LastForceRefresh()) < config.Opts.ForceRefreshInterval() {
		interval := int(config.Opts.ForceRefreshInterval().Minutes())
		sess.SetErrorMessage(printer.Plural("alert.too_many_feeds_refresh", interval, interval))
	} else {
		userID := request.UserID(r)
		// We allow the end-user to force refresh all its feeds
		// without taking into consideration the number of errors.
		batchBuilder := h.store.NewBatchBuilder()
		batchBuilder.WithoutDisabledFeeds()
		batchBuilder.WithUserID(userID)
		batchBuilder.WithLimitPerHost(config.Opts.PollingLimitPerHost())

		jobs, err := batchBuilder.FetchJobs()
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		slog.Info(
			"Triggered a manual refresh of all feeds from the web ui",
			slog.Int64("user_id", userID),
			slog.Int("nb_jobs", len(jobs)),
		)

		go h.pool.Push(jobs)

		sess.MarkForceRefreshed()
		sess.SetSuccessMessage(printer.Print("alert.background_feed_refresh"))
	}

	response.HTMLRedirect(w, r, h.routePath("/feeds"))
}
