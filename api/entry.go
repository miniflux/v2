// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response/json"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/storage"
)

// GetFeedEntry is the API handler to get a single feed entry.
func (c *Controller) GetFeedEntry(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	entryID, err := request.IntParam(r, "entryID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	ctx := context.New(r)
	userID := ctx.UserID()

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		json.ServerError(w, errors.New("Unable to fetch this entry from the database"))
		return
	}

	if entry == nil {
		json.NotFound(w, errors.New("Entry not found"))
		return
	}

	json.OK(w, r, entry)
}

// GetEntry is the API handler to get a single entry.
func (c *Controller) GetEntry(w http.ResponseWriter, r *http.Request) {
	entryID, err := request.IntParam(r, "entryID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(context.New(r).UserID())
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		json.ServerError(w, errors.New("Unable to fetch this entry from the database"))
		return
	}

	if entry == nil {
		json.NotFound(w, errors.New("Entry not found"))
		return
	}

	json.OK(w, r, entry)
}

// GetFeedEntries is the API handler to get all feed entries.
func (c *Controller) GetFeedEntries(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	status := request.QueryParam(r, "status", "")
	if status != "" {
		if err := model.ValidateEntryStatus(status); err != nil {
			json.BadRequest(w, err)
			return
		}
	}

	order := request.QueryParam(r, "order", model.DefaultSortingOrder)
	if err := model.ValidateEntryOrder(order); err != nil {
		json.BadRequest(w, err)
		return
	}

	direction := request.QueryParam(r, "direction", model.DefaultSortingDirection)
	if err := model.ValidateDirection(direction); err != nil {
		json.BadRequest(w, err)
		return
	}

	limit := request.QueryIntParam(r, "limit", 100)
	offset := request.QueryIntParam(r, "offset", 0)
	if err := model.ValidateRange(offset, limit); err != nil {
		json.BadRequest(w, err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(context.New(r).UserID())
	builder.WithFeedID(feedID)
	builder.WithStatus(status)
	builder.WithOrder(order)
	builder.WithDirection(direction)
	builder.WithOffset(offset)
	builder.WithLimit(limit)
	configureFilters(builder, r)

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, errors.New("Unable to fetch the list of entries"))
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		json.ServerError(w, errors.New("Unable to count the number of entries"))
		return
	}

	json.OK(w, r, &entriesResponse{Total: count, Entries: entries})
}

// GetEntries is the API handler to fetch entries.
func (c *Controller) GetEntries(w http.ResponseWriter, r *http.Request) {
	status := request.QueryParam(r, "status", "")
	if status != "" {
		if err := model.ValidateEntryStatus(status); err != nil {
			json.BadRequest(w, err)
			return
		}
	}

	order := request.QueryParam(r, "order", model.DefaultSortingOrder)
	if err := model.ValidateEntryOrder(order); err != nil {
		json.BadRequest(w, err)
		return
	}

	direction := request.QueryParam(r, "direction", model.DefaultSortingDirection)
	if err := model.ValidateDirection(direction); err != nil {
		json.BadRequest(w, err)
		return
	}

	limit := request.QueryIntParam(r, "limit", 100)
	offset := request.QueryIntParam(r, "offset", 0)
	if err := model.ValidateRange(offset, limit); err != nil {
		json.BadRequest(w, err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(context.New(r).UserID())
	builder.WithStatus(status)
	builder.WithOrder(order)
	builder.WithDirection(direction)
	builder.WithOffset(offset)
	builder.WithLimit(limit)
	configureFilters(builder, r)

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, errors.New("Unable to fetch the list of entries"))
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		json.ServerError(w, errors.New("Unable to count the number of entries"))
		return
	}

	json.OK(w, r, &entriesResponse{Total: count, Entries: entries})
}

// SetEntryStatus is the API handler to change the status of entries.
func (c *Controller) SetEntryStatus(w http.ResponseWriter, r *http.Request) {
	entryIDs, status, err := decodeEntryStatusPayload(r.Body)
	if err != nil {
		json.BadRequest(w, errors.New("Invalid JSON payload"))
		return
	}

	if err := model.ValidateEntryStatus(status); err != nil {
		json.BadRequest(w, err)
		return
	}

	if err := c.store.SetEntriesStatus(context.New(r).UserID(), entryIDs, status); err != nil {
		json.ServerError(w, errors.New("Unable to change entries status"))
		return
	}

	json.NoContent(w)
}

// ToggleBookmark is the API handler to toggle bookmark status.
func (c *Controller) ToggleBookmark(w http.ResponseWriter, r *http.Request) {
	entryID, err := request.IntParam(r, "entryID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	if err := c.store.ToggleBookmark(context.New(r).UserID(), entryID); err != nil {
		json.ServerError(w, errors.New("Unable to toggle bookmark value"))
		return
	}

	json.NoContent(w)
}

func configureFilters(builder *storage.EntryQueryBuilder, r *http.Request) {
	beforeEntryID := request.QueryInt64Param(r, "before_entry_id", 0)
	if beforeEntryID != 0 {
		builder.BeforeEntryID(beforeEntryID)
	}

	afterEntryID := request.QueryInt64Param(r, "after_entry_id", 0)
	if afterEntryID != 0 {
		builder.AfterEntryID(afterEntryID)
	}

	beforeTimestamp := request.QueryInt64Param(r, "before", 0)
	if beforeTimestamp != 0 {
		builder.BeforeDate(time.Unix(beforeTimestamp, 0))
	}

	afterTimestamp := request.QueryInt64Param(r, "after", 0)
	if afterTimestamp != 0 {
		builder.AfterDate(time.Unix(afterTimestamp, 0))
	}

	if request.HasQueryParam(r, "starred") {
		builder.WithStarred()
	}

	searchQuery := request.QueryParam(r, "search", "")
	if searchQuery != "" {
		builder.WithSearchQuery(searchQuery)
	}
}
