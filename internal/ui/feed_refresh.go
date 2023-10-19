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
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
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

	user, err := h.store.UserByID(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	jobs, err := h.store.NewUserBatch(userID, h.store.CountFeeds(userID))
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

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	html.OK(w, r, view.Render("feed_background_refresh"))
}
