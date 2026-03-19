// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	feedHandler "miniflux.app/v2/internal/reader/handler"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) createFeedHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	var feedCreationRequest model.FeedCreationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&feedCreationRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	// Make the feed category optional for clients who don't support categories.
	if feedCreationRequest.CategoryID == 0 {
		category, err := h.store.FirstCategory(userID)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
		feedCreationRequest.CategoryID = category.ID
	}

	if validationErr := validator.ValidateFeedCreation(h.store, userID, &feedCreationRequest); validationErr != nil {
		response.JSONBadRequest(w, r, validationErr.Error())
		return
	}

	feed, localizedError := feedHandler.CreateFeed(h.store, userID, &feedCreationRequest)
	if localizedError != nil {
		response.JSONServerError(w, r, localizedError.Error())
		return
	}

	response.JSONCreated(w, r, &feedCreationResponse{FeedID: feed.ID})
}

func (h *handler) refreshFeedHandler(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if feedID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	userID := request.UserID(r)
	if !h.store.FeedExists(userID, feedID) {
		response.JSONNotFound(w, r)
		return
	}

	localizedError := feedHandler.RefreshFeed(h.store, userID, feedID, false)
	if localizedError != nil {
		response.JSONServerError(w, r, localizedError.Error())
		return
	}

	response.NoContent(w, r)
}

func (h *handler) refreshAllFeedsHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	batchBuilder := h.store.NewBatchBuilder()
	batchBuilder.WithErrorLimit(config.Opts.PollingParsingErrorLimit())
	batchBuilder.WithoutDisabledFeeds()
	batchBuilder.WithNextCheckExpired()
	batchBuilder.WithUserID(userID)
	batchBuilder.WithLimitPerHost(config.Opts.PollingLimitPerHost())

	jobs, err := batchBuilder.FetchJobs()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	slog.Info(
		"Triggered a manual refresh of all feeds from the API",
		slog.Int64("user_id", userID),
		slog.Int("nb_jobs", len(jobs)),
	)

	go h.pool.Push(jobs)

	response.NoContent(w, r)
}

func (h *handler) updateFeedHandler(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if feedID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	var feedModificationRequest model.FeedModificationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&feedModificationRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	userID := request.UserID(r)
	originalFeed, err := h.store.FeedByID(userID, feedID)
	if err != nil {
		response.JSONNotFound(w, r)
		return
	}

	if originalFeed == nil {
		response.JSONNotFound(w, r)
		return
	}

	if validationErr := validator.ValidateFeedModification(h.store, userID, originalFeed.ID, &feedModificationRequest); validationErr != nil {
		response.JSONBadRequest(w, r, validationErr.Error())
		return
	}

	feedModificationRequest.Patch(originalFeed)
	originalFeed.ResetErrorCounter()
	if err := h.store.UpdateFeed(originalFeed); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	originalFeed, err = h.store.FeedByID(userID, feedID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSONCreated(w, r, originalFeed)
}

func (h *handler) markFeedAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	feedID := request.RouteInt64Param(r, "feedID")
	if feedID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	if !h.store.FeedExists(userID, feedID) {
		response.JSONNotFound(w, r)
		return
	}

	if err := h.store.MarkFeedAsRead(userID, feedID, time.Now()); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}

func (h *handler) getCategoryFeedsHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	categoryID := request.RouteInt64Param(r, "categoryID")
	if categoryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid category ID"))
		return
	}

	category, err := h.store.Category(userID, categoryID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if category == nil {
		response.JSONNotFound(w, r)
		return
	}

	feeds, err := h.store.FeedsByCategoryWithCounters(userID, categoryID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSON(w, r, feeds)
}

func (h *handler) getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.store.Feeds(request.UserID(r))
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSON(w, r, feeds)
}

func (h *handler) fetchCountersHandler(w http.ResponseWriter, r *http.Request) {
	counters, err := h.store.FetchCounters(request.UserID(r))
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSON(w, r, counters)
}

func (h *handler) getFeedHandler(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if feedID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	feed, err := h.store.FeedByID(request.UserID(r), feedID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if feed == nil {
		response.JSONNotFound(w, r)
		return
	}

	response.JSON(w, r, feed)
}

func (h *handler) removeFeedHandler(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if feedID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	userID := request.UserID(r)
	if !h.store.FeedExists(userID, feedID) {
		response.JSONNotFound(w, r)
		return
	}

	if err := h.store.RemoveFeed(userID, feedID); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}
