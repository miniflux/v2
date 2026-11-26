// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/locale"
)

const (
	bulkFeedActionRefresh      = "refresh"
	bulkFeedActionMarkAsRead   = "mark_as_read"
	bulkFeedActionEnable       = "enable"
	bulkFeedActionDisable      = "disable"
	bulkFeedActionMoveCategory = "move"
	bulkFeedActionRemove       = "remove"
)

// bulkUpdateFeeds applies an action to the selected feeds of the "/feeds" page.
func (h *handler) bulkUpdateFeeds(w http.ResponseWriter, r *http.Request) {
	if h.applyBulkFeedAction(w, r) {
		response.HTMLRedirect(w, r, h.routePath("/feeds"))
	}
}

// bulkUpdateCategoryFeeds applies an action to the selected feeds of the "/category/{categoryID}/feeds" page.
func (h *handler) bulkUpdateCategoryFeeds(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	if h.applyBulkFeedAction(w, r) {
		response.HTMLRedirect(w, r, h.routePath("/category/%d/feeds", categoryID))
	}
}

// applyBulkFeedAction runs the requested bulk action and returns true when the caller should redirect.
// It returns false when an HTTP response has already been written.
func (h *handler) applyBulkFeedAction(w http.ResponseWriter, r *http.Request) bool {
	userID := request.UserID(r)
	sess := request.WebSession(r)
	printer := locale.NewPrinter(sess.Language())

	if err := r.ParseForm(); err != nil {
		response.HTMLBadRequest(w, r, err)
		return false
	}

	feedIDs := parseFeedIDs(r.Form["feed_ids"])
	if len(feedIDs) == 0 {
		sess.SetErrorMessage(printer.Print("alert.bulk_feeds_none_selected"))
		return true
	}

	action := r.FormValue("action")
	slog.Info("Bulk feed action requested from the web ui",
		slog.Int64("user_id", userID),
		slog.String("action", action),
		slog.Int("nb_feeds", len(feedIDs)),
	)

	var err error
	switch action {
	case bulkFeedActionRefresh:
		// Avoid accidental and excessive refreshes, like the "refresh all feeds" action.
		if time.Since(sess.LastForceRefresh()) < config.Opts.ForceRefreshInterval() {
			interval := int(config.Opts.ForceRefreshInterval().Minutes())
			sess.SetErrorMessage(printer.Plural("alert.too_many_feeds_refresh", interval, interval))
			return true
		}

		jobs, fetchErr := h.store.NewBatchBuilder().
			WithoutDisabledFeeds().
			WithUserID(userID).
			WithFeedIDs(feedIDs).
			WithLimitPerHost(config.Opts.PollingLimitPerHost()).
			FetchJobs()
		if fetchErr != nil {
			response.HTMLServerError(w, r, fetchErr)
			return false
		}

		go h.pool.Push(jobs)

		sess.MarkForceRefreshed()
		sess.SetSuccessMessage(printer.Print("alert.background_feed_refresh"))
		return true
	case bulkFeedActionMarkAsRead:
		err = h.store.MarkFeedsAsRead(userID, feedIDs)
	case bulkFeedActionEnable:
		err = h.store.UpdateFeedsDisabled(userID, feedIDs, false)
	case bulkFeedActionDisable:
		err = h.store.UpdateFeedsDisabled(userID, feedIDs, true)
	case bulkFeedActionMoveCategory:
		categoryID := request.FormInt64Value(r, "category_id")
		exists, existsErr := h.store.CategoryIDExists(userID, categoryID)
		if existsErr != nil {
			response.HTMLServerError(w, r, existsErr)
			return false
		}
		if !exists {
			sess.SetErrorMessage(printer.Print("error.category_not_found"))
			return true
		}
		err = h.store.UpdateFeedsCategory(userID, categoryID, feedIDs)
	case bulkFeedActionRemove:
		err = h.store.RemoveFeeds(userID, feedIDs)
	default:
		response.HTMLBadRequest(w, r, nil)
		return false
	}

	if err != nil {
		response.HTMLServerError(w, r, err)
		return false
	}

	sess.SetSuccessMessage(printer.Print("alert.bulk_feeds_updated"))
	return true
}

// parseFeedIDs converts the submitted form values into a list of unique, positive feed IDs.
func parseFeedIDs(values []string) []int64 {
	seen := make(map[int64]struct{}, len(values))
	feedIDs := make([]int64, 0, len(values))
	for _, value := range values {
		feedID, err := strconv.ParseInt(value, 10, 64)
		if err != nil || feedID <= 0 {
			continue
		}
		if _, found := seen[feedID]; found {
			continue
		}
		seen[feedID] = struct{}{}
		feedIDs = append(feedIDs, feedID)
	}
	return feedIDs
}
