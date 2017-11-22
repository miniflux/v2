// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/api/payload"
	"github.com/miniflux/miniflux2/server/core"
)

// GetEntry is the API handler to get a single feed entry.
func (c *Controller) GetEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	userID := ctx.GetUserID()
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(userID, ctx.GetUserTimezone())
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		response.Json().ServerError(errors.New("Unable to fetch this entry from the database"))
		return
	}

	if entry == nil {
		response.Json().NotFound(errors.New("Entry not found"))
		return
	}

	response.Json().Standard(entry)
}

// GetFeedEntries is the API handler to get all feed entries.
func (c *Controller) GetFeedEntries(ctx *core.Context, request *core.Request, response *core.Response) {
	userID := ctx.GetUserID()
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	status := request.QueryStringParam("status", "")
	if status != "" {
		if err := model.ValidateEntryStatus(status); err != nil {
			response.Json().BadRequest(err)
			return
		}
	}

	order := request.QueryStringParam("order", "id")
	if err := model.ValidateEntryOrder(order); err != nil {
		response.Json().BadRequest(err)
		return
	}

	direction := request.QueryStringParam("direction", "desc")
	if err := model.ValidateDirection(direction); err != nil {
		response.Json().BadRequest(err)
		return
	}

	limit := request.QueryIntegerParam("limit", 100)
	offset := request.QueryIntegerParam("offset", 0)

	builder := c.store.GetEntryQueryBuilder(userID, ctx.GetUserTimezone())
	builder.WithFeedID(feedID)
	builder.WithStatus(status)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(model.DefaultSortingDirection)
	builder.WithOffset(offset)
	builder.WithLimit(limit)

	entries, err := builder.GetEntries()
	if err != nil {
		response.Json().ServerError(errors.New("Unable to fetch the list of entries"))
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		response.Json().ServerError(errors.New("Unable to count the number of entries"))
		return
	}

	response.Json().Standard(&payload.EntriesResponse{Total: count, Entries: entries})
}

// SetEntryStatus is the API handler to change the status of an entry.
func (c *Controller) SetEntryStatus(ctx *core.Context, request *core.Request, response *core.Response) {
	userID := ctx.GetUserID()

	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	status, err := payload.DecodeEntryStatusPayload(request.Body())
	if err != nil {
		response.Json().BadRequest(errors.New("Invalid JSON payload"))
		return
	}

	if err := model.ValidateEntryStatus(status); err != nil {
		response.Json().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(userID, ctx.GetUserTimezone())
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		response.Json().ServerError(errors.New("Unable to fetch this entry from the database"))
		return
	}

	if entry == nil {
		response.Json().NotFound(errors.New("Entry not found"))
		return
	}

	if err := c.store.SetEntriesStatus(userID, []int64{entry.ID}, status); err != nil {
		response.Json().ServerError(errors.New("Unable to change entry status"))
		return
	}

	entry, err = builder.GetEntry()
	if err != nil {
		response.Json().ServerError(errors.New("Unable to fetch this entry from the database"))
		return
	}

	response.Json().Standard(entry)
}
