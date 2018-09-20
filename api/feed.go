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

// CreateFeed is the API handler to create a new feed.
func (c *Controller) CreateFeed(w http.ResponseWriter, r *http.Request) {
	feedInfo, err := decodeFeedCreationPayload(r.Body)
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	if feedInfo.FeedURL == "" {
		json.BadRequest(w, errors.New("The feed_url is required"))
		return
	}

	if feedInfo.CategoryID <= 0 {
		json.BadRequest(w, errors.New("The category_id is required"))
		return
	}

	userID := request.UserID(r)

	if c.store.FeedURLExists(userID, feedInfo.FeedURL) {
		json.BadRequest(w, errors.New("This feed_url already exists"))
		return
	}

	if !c.store.CategoryExists(userID, feedInfo.CategoryID) {
		json.BadRequest(w, errors.New("This category_id doesn't exists or doesn't belongs to this user"))
		return
	}

	feed, err := c.feedHandler.CreateFeed(
		userID,
		feedInfo.CategoryID,
		feedInfo.FeedURL,
		feedInfo.Crawler,
		feedInfo.UserAgent,
		feedInfo.Username,
		feedInfo.Password,
	)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	type result struct {
		FeedID int64 `json:"feed_id"`
	}

	json.Created(w, &result{FeedID: feed.ID})
}

// RefreshFeed is the API handler to refresh a feed.
func (c *Controller) RefreshFeed(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	userID := request.UserID(r)

	if !c.store.FeedExists(userID, feedID) {
		json.NotFound(w, errors.New("Unable to find this feed"))
		return
	}

	err = c.feedHandler.RefreshFeed(userID, feedID)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	json.NoContent(w)
}

// UpdateFeed is the API handler that is used to update a feed.
func (c *Controller) UpdateFeed(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	feedChanges, err := decodeFeedModificationPayload(r.Body)
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	userID := request.UserID(r)

	originalFeed, err := c.store.FeedByID(userID, feedID)
	if err != nil {
		json.NotFound(w, errors.New("Unable to find this feed"))
		return
	}

	if originalFeed == nil {
		json.NotFound(w, errors.New("Feed not found"))
		return
	}

	feedChanges.Update(originalFeed)

	if !c.store.CategoryExists(userID, originalFeed.Category.ID) {
		json.BadRequest(w, errors.New("This category_id doesn't exists or doesn't belongs to this user"))
		return
	}

	if err := c.store.UpdateFeed(originalFeed); err != nil {
		json.ServerError(w, err)
		return
	}

	originalFeed, err = c.store.FeedByID(userID, feedID)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	json.Created(w, originalFeed)
}

// GetFeeds is the API handler that get all feeds that belongs to the given user.
func (c *Controller) GetFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := c.store.Feeds(request.UserID(r))
	if err != nil {
		json.ServerError(w, err)
		return
	}

	json.OK(w, r, feeds)
}

// GetFeed is the API handler to get a feed.
func (c *Controller) GetFeed(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	feed, err := c.store.FeedByID(request.UserID(r), feedID)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	if feed == nil {
		json.NotFound(w, errors.New("Feed not found"))
		return
	}

	json.OK(w, r, feed)
}

// RemoveFeed is the API handler to remove a feed.
func (c *Controller) RemoveFeed(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	userID := request.UserID(r)

	if !c.store.FeedExists(userID, feedID) {
		json.NotFound(w, errors.New("Feed not found"))
		return
	}

	if err := c.store.RemoveFeed(userID, feedID); err != nil {
		json.ServerError(w, err)
		return
	}

	json.NoContent(w)
}
