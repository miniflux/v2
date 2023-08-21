// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/api"

import (
	json_parser "encoding/json"
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/model"
	feedHandler "miniflux.app/reader/handler"
	"miniflux.app/validator"
)

func (h *handler) createFeed(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	var feedCreationRequest model.FeedCreationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&feedCreationRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateFeedCreation(h.store, userID, &feedCreationRequest); validationErr != nil {
		json.BadRequest(w, r, validationErr.Error())
		return
	}

	feed, err := feedHandler.CreateFeed(h.store, userID, &feedCreationRequest)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, &feedCreationResponse{FeedID: feed.ID})
}

func (h *handler) refreshFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	userID := request.UserID(r)

	if !h.store.FeedExists(userID, feedID) {
		json.NotFound(w, r)
		return
	}

	err := feedHandler.RefreshFeed(h.store, userID, feedID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) refreshAllFeeds(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	jobs, err := h.store.NewUserBatch(userID, h.store.CountFeeds(userID))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	go func() {
		h.pool.Push(jobs)
	}()

	json.NoContent(w, r)
}

func (h *handler) updateFeed(w http.ResponseWriter, r *http.Request) {
	var feedModificationRequest model.FeedModificationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&feedModificationRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	userID := request.UserID(r)
	feedID := request.RouteInt64Param(r, "feedID")

	originalFeed, err := h.store.FeedByID(userID, feedID)
	if err != nil {
		json.NotFound(w, r)
		return
	}

	if originalFeed == nil {
		json.NotFound(w, r)
		return
	}

	if validationErr := validator.ValidateFeedModification(h.store, userID, &feedModificationRequest); validationErr != nil {
		json.BadRequest(w, r, validationErr.Error())
		return
	}

	feedModificationRequest.Patch(originalFeed)
	if err := h.store.UpdateFeed(originalFeed); err != nil {
		json.ServerError(w, r, err)
		return
	}

	originalFeed, err = h.store.FeedByID(userID, feedID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, originalFeed)
}

func (h *handler) markFeedAsRead(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	userID := request.UserID(r)

	feed, err := h.store.FeedByID(userID, feedID)
	if err != nil {
		json.NotFound(w, r)
		return
	}

	if feed == nil {
		json.NotFound(w, r)
		return
	}

	if err := h.store.MarkFeedAsRead(userID, feedID, time.Now()); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) getCategoryFeeds(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	categoryID := request.RouteInt64Param(r, "categoryID")

	category, err := h.store.Category(userID, categoryID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if category == nil {
		json.NotFound(w, r)
		return
	}

	feeds, err := h.store.FeedsByCategoryWithCounters(userID, categoryID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, feeds)
}

func (h *handler) getFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.store.Feeds(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, feeds)
}

func (h *handler) fetchCounters(w http.ResponseWriter, r *http.Request) {
	counters, err := h.store.FetchCounters(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, counters)
}

func (h *handler) getFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	feed, err := h.store.FeedByID(request.UserID(r), feedID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if feed == nil {
		json.NotFound(w, r)
		return
	}

	json.OK(w, r, feed)
}

func (h *handler) removeFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	userID := request.UserID(r)

	if !h.store.FeedExists(userID, feedID) {
		json.NotFound(w, r)
		return
	}

	if err := h.store.RemoveFeed(userID, feedID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}
