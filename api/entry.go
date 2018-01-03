// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"

	"github.com/miniflux/miniflux/http/handler"
	"github.com/miniflux/miniflux/model"
)

// GetFeedEntry is the API handler to get a single feed entry.
func (c *Controller) GetFeedEntry(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch this entry from the database"))
		return
	}

	if entry == nil {
		response.JSON().NotFound(errors.New("Entry not found"))
		return
	}

	response.JSON().Standard(entry)
}

// GetEntry is the API handler to get a single entry.
func (c *Controller) GetEntry(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch this entry from the database"))
		return
	}

	if entry == nil {
		response.JSON().NotFound(errors.New("Entry not found"))
		return
	}

	response.JSON().Standard(entry)
}

// GetFeedEntries is the API handler to get all feed entries.
func (c *Controller) GetFeedEntries(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	status := request.QueryStringParam("status", "")
	if status != "" {
		if err := model.ValidateEntryStatus(status); err != nil {
			response.JSON().BadRequest(err)
			return
		}
	}

	order := request.QueryStringParam("order", model.DefaultSortingOrder)
	if err := model.ValidateEntryOrder(order); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	direction := request.QueryStringParam("direction", model.DefaultSortingDirection)
	if err := model.ValidateDirection(direction); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	limit := request.QueryIntegerParam("limit", 100)
	offset := request.QueryIntegerParam("offset", 0)
	if err := model.ValidateRange(offset, limit); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithFeedID(feedID)
	builder.WithStatus(status)
	builder.WithOrder(order)
	builder.WithDirection(direction)
	builder.WithOffset(offset)
	builder.WithLimit(limit)

	entries, err := builder.GetEntries()
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch the list of entries"))
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to count the number of entries"))
		return
	}

	response.JSON().Standard(&entriesResponse{Total: count, Entries: entries})
}

// GetEntries is the API handler to fetch entries.
func (c *Controller) GetEntries(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()

	status := request.QueryStringParam("status", "")
	if status != "" {
		if err := model.ValidateEntryStatus(status); err != nil {
			response.JSON().BadRequest(err)
			return
		}
	}

	order := request.QueryStringParam("order", model.DefaultSortingOrder)
	if err := model.ValidateEntryOrder(order); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	direction := request.QueryStringParam("direction", model.DefaultSortingDirection)
	if err := model.ValidateDirection(direction); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	limit := request.QueryIntegerParam("limit", 100)
	offset := request.QueryIntegerParam("offset", 0)
	if err := model.ValidateRange(offset, limit); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithStatus(status)
	builder.WithOrder(order)
	builder.WithDirection(direction)
	builder.WithOffset(offset)
	builder.WithLimit(limit)

	entries, err := builder.GetEntries()
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch the list of entries"))
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to count the number of entries"))
		return
	}

	response.JSON().Standard(&entriesResponse{Total: count, Entries: entries})
}

// SetEntryStatus is the API handler to change the status of entries.
func (c *Controller) SetEntryStatus(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()

	entryIDs, status, err := decodeEntryStatusPayload(request.Body())
	if err != nil {
		response.JSON().BadRequest(errors.New("Invalid JSON payload"))
		return
	}

	if err := model.ValidateEntryStatus(status); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if err := c.store.SetEntriesStatus(userID, entryIDs, status); err != nil {
		response.JSON().ServerError(errors.New("Unable to change entries status"))
		return
	}

	response.JSON().NoContent()
}

// ToggleBookmark is the API handler to toggle bookmark status.
func (c *Controller) ToggleBookmark(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if err := c.store.ToggleBookmark(userID, entryID); err != nil {
		response.JSON().ServerError(errors.New("Unable to toggle bookmark value"))
		return
	}

	response.JSON().NoContent()
}
