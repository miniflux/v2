// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	feedHandler "miniflux.app/v2/internal/reader/handler"
	"miniflux.app/v2/internal/ui/session"
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

	html.Redirect(w, r, route.Path(h.router, "feedEntries", "feedID", feedID))
}

func (h *handler) refreshAllFeeds(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	printer := locale.NewPrinter(request.UserLanguage(r))
	sess := session.New(h.store, request.SessionID(r))

	// Avoid accidental and excessive refreshes.
	if time.Now().UTC().Unix()-request.LastForceRefresh(r) < int64(config.Opts.ForceRefreshInterval())*60 {
		time := config.Opts.ForceRefreshInterval()
		sess.NewFlashErrorMessage(printer.Plural("alert.too_many_feeds_refresh", time, time))
	} else {
		// We allow the end-user to force refresh all its feeds
		// without taking into consideration the number of errors.
		batchBuilder := h.store.NewBatchBuilder()
		batchBuilder.WithoutDisabledFeeds()
		batchBuilder.WithUserID(userID)

		jobs, err := batchBuilder.FetchJobs()
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		slog.Info(
			"Triggered a manual refresh of all feeds from the web ui",
			slog.Int64("user_id", userID),
			slog.Int("nb_jobs", len(jobs)),
		)

		go h.pool.Push(jobs)

		sess.SetLastForceRefresh()
		sess.NewFlashMessage(printer.Print("alert.background_feed_refresh"))
	}

	html.Redirect(w, r, route.Path(h.router, "feeds"))
}
