// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"errors"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
)

func (h *handler) createFeed(w http.ResponseWriter, r *http.Request) {
	feedInfo, err := decodeFeedCreationPayload(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if feedInfo.FeedURL == "" {
		json.BadRequest(w, r, errors.New("The feed_url is required"))
		return
	}

	if feedInfo.CategoryID <= 0 {
		json.BadRequest(w, r, errors.New("The category_id is required"))
		return
	}

	userID := request.UserID(r)

	if h.store.FeedURLExists(userID, feedInfo.FeedURL) {
		json.BadRequest(w, r, errors.New("This feed_url already exists"))
		return
	}

	if !h.store.CategoryExists(userID, feedInfo.CategoryID) {
		json.BadRequest(w, r, errors.New("This category_id doesn't exists or doesn't belongs to this user"))
		return
	}

	feed, err := h.feedHandler.CreateFeed(
		userID,
		feedInfo.CategoryID,
		feedInfo.FeedURL,
		feedInfo.Crawler,
		feedInfo.UserAgent,
		feedInfo.Username,
		feedInfo.Password,
		feedInfo.ScraperRules,
		feedInfo.RewriteRules,
		feedInfo.BlocklistRules,
		feedInfo.KeeplistRules,
		feedInfo.FetchViaProxy,
	)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	type result struct {
		FeedID int64 `json:"feed_id"`
	}

	json.Created(w, r, &result{FeedID: feed.ID})
}

func (h *handler) refreshFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	userID := request.UserID(r)

	if !h.store.FeedExists(userID, feedID) {
		json.NotFound(w, r)
		return
	}

	err := h.feedHandler.RefreshFeed(userID, feedID)
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
	feedID := request.RouteInt64Param(r, "feedID")
	feedChanges, err := decodeFeedModificationPayload(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	userID := request.UserID(r)

	originalFeed, err := h.store.FeedByID(userID, feedID)
	if err != nil {
		json.NotFound(w, r)
		return
	}

	if originalFeed == nil {
		json.NotFound(w, r)
		return
	}

	feedChanges.Update(originalFeed)

	if !h.store.CategoryExists(userID, originalFeed.Category.ID) {
		json.BadRequest(w, r, errors.New("This category_id doesn't exists or doesn't belongs to this user"))
		return
	}

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

func (h *handler) getFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.store.Feeds(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, feeds)
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
