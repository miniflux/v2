// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"

	"github.com/miniflux/miniflux/http/handler"
)

// CreateFeed is the API handler to create a new feed.
func (c *Controller) CreateFeed(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	feedURL, categoryID, err := decodeFeedCreationPayload(request.Body())
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if feedURL == "" {
		response.JSON().BadRequest(errors.New("The feed_url is required"))
		return
	}

	if categoryID <= 0 {
		response.JSON().BadRequest(errors.New("The category_id is required"))
		return
	}

	if c.store.FeedURLExists(userID, feedURL) {
		response.JSON().BadRequest(errors.New("This feed_url already exists"))
		return
	}

	if !c.store.CategoryExists(userID, categoryID) {
		response.JSON().BadRequest(errors.New("This category_id doesn't exists or doesn't belongs to this user"))
		return
	}

	feed, err := c.feedHandler.CreateFeed(userID, categoryID, feedURL, false)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to create this feed"))
		return
	}

	type result struct {
		FeedID int64 `json:"feed_id"`
	}

	response.JSON().Created(&result{FeedID: feed.ID})
}

// RefreshFeed is the API handler to refresh a feed.
func (c *Controller) RefreshFeed(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if !c.store.FeedExists(userID, feedID) {
		response.JSON().NotFound(errors.New("Unable to find this feed"))
		return
	}

	err = c.feedHandler.RefreshFeed(userID, feedID)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to refresh this feed"))
		return
	}

	response.JSON().NoContent()
}

// UpdateFeed is the API handler that is used to update a feed.
func (c *Controller) UpdateFeed(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	newFeed, err := decodeFeedModificationPayload(request.Body())
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if newFeed.Category != nil && newFeed.Category.ID != 0 && !c.store.CategoryExists(userID, newFeed.Category.ID) {
		response.JSON().BadRequest(errors.New("This category_id doesn't exists or doesn't belongs to this user"))
		return
	}

	originalFeed, err := c.store.FeedByID(userID, feedID)
	if err != nil {
		response.JSON().NotFound(errors.New("Unable to find this feed"))
		return
	}

	if originalFeed == nil {
		response.JSON().NotFound(errors.New("Feed not found"))
		return
	}

	originalFeed.Merge(newFeed)
	if err := c.store.UpdateFeed(originalFeed); err != nil {
		response.JSON().ServerError(errors.New("Unable to update this feed"))
		return
	}

	originalFeed, err = c.store.FeedByID(userID, feedID)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch this feed"))
		return
	}

	response.JSON().Created(originalFeed)
}

// GetFeeds is the API handler that get all feeds that belongs to the given user.
func (c *Controller) GetFeeds(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	feeds, err := c.store.Feeds(ctx.UserID())
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch feeds from the database"))
		return
	}

	response.JSON().Standard(feeds)
}

// GetFeed is the API handler to get a feed.
func (c *Controller) GetFeed(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	feed, err := c.store.FeedByID(userID, feedID)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch this feed"))
		return
	}

	if feed == nil {
		response.JSON().NotFound(errors.New("Feed not found"))
		return
	}

	response.JSON().Standard(feed)
}

// RemoveFeed is the API handler to remove a feed.
func (c *Controller) RemoveFeed(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if !c.store.FeedExists(userID, feedID) {
		response.JSON().NotFound(errors.New("Feed not found"))
		return
	}

	if err := c.store.RemoveFeed(userID, feedID); err != nil {
		response.JSON().ServerError(errors.New("Unable to remove this feed"))
		return
	}

	response.JSON().NoContent()
}
