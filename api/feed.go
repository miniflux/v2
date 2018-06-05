// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response/json"
)

// CreateFeed is the API handler to create a new feed.
func (c *Controller) CreateFeed(w http.ResponseWriter, r *http.Request) {
	feedURL, categoryID, err := decodeFeedCreationPayload(r.Body)
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	if feedURL == "" {
		json.BadRequest(w, errors.New("The feed_url is required"))
		return
	}

	if categoryID <= 0 {
		json.BadRequest(w, errors.New("The category_id is required"))
		return
	}

	ctx := context.New(r)
	userID := ctx.UserID()

	if c.store.FeedURLExists(userID, feedURL) {
		json.BadRequest(w, errors.New("This feed_url already exists"))
		return
	}

	if !c.store.CategoryExists(userID, categoryID) {
		json.BadRequest(w, errors.New("This category_id doesn't exists or doesn't belongs to this user"))
		return
	}

	feed, err := c.feedHandler.CreateFeed(userID, categoryID, feedURL, false)
	if err != nil {
		json.ServerError(w, errors.New("Unable to create this feed"))
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

	ctx := context.New(r)
	userID := ctx.UserID()

	if !c.store.FeedExists(userID, feedID) {
		json.NotFound(w, errors.New("Unable to find this feed"))
		return
	}

	err = c.feedHandler.RefreshFeed(userID, feedID)
	if err != nil {
		json.ServerError(w, errors.New("Unable to refresh this feed"))
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

	newFeed, err := decodeFeedModificationPayload(r.Body)
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	ctx := context.New(r)
	userID := ctx.UserID()

	if newFeed.Category != nil && newFeed.Category.ID != 0 && !c.store.CategoryExists(userID, newFeed.Category.ID) {
		json.BadRequest(w, errors.New("This category_id doesn't exists or doesn't belongs to this user"))
		return
	}

	originalFeed, err := c.store.FeedByID(userID, feedID)
	if err != nil {
		json.NotFound(w, errors.New("Unable to find this feed"))
		return
	}

	if originalFeed == nil {
		json.NotFound(w, errors.New("Feed not found"))
		return
	}

	originalFeed.Merge(newFeed)
	if err := c.store.UpdateFeed(originalFeed); err != nil {
		json.ServerError(w, errors.New("Unable to update this feed"))
		return
	}

	originalFeed, err = c.store.FeedByID(userID, feedID)
	if err != nil {
		json.ServerError(w, errors.New("Unable to fetch this feed"))
		return
	}

	json.Created(w, originalFeed)
}

// GetFeeds is the API handler that get all feeds that belongs to the given user.
func (c *Controller) GetFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := c.store.Feeds(context.New(r).UserID())
	if err != nil {
		json.ServerError(w, errors.New("Unable to fetch feeds from the database"))
		return
	}

	json.OK(w, feeds)
}

// GetFeed is the API handler to get a feed.
func (c *Controller) GetFeed(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	feed, err := c.store.FeedByID(context.New(r).UserID(), feedID)
	if err != nil {
		json.ServerError(w, errors.New("Unable to fetch this feed"))
		return
	}

	if feed == nil {
		json.NotFound(w, errors.New("Feed not found"))
		return
	}

	json.OK(w, feed)
}

// RemoveFeed is the API handler to remove a feed.
func (c *Controller) RemoveFeed(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	ctx := context.New(r)
	userID := ctx.UserID()

	if !c.store.FeedExists(userID, feedID) {
		json.NotFound(w, errors.New("Feed not found"))
		return
	}

	if err := c.store.RemoveFeed(userID, feedID); err != nil {
		json.ServerError(w, errors.New("Unable to remove this feed"))
		return
	}

	json.NoContent(w)
}
