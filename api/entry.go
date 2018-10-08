// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"errors"
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/model"
	"miniflux.app/storage"
)

// GetFeedEntry is the API handler to get a single feed entry.
func (c *Controller) GetFeedEntry(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := c.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if entry == nil {
		json.NotFound(w, r)
		return
	}

	json.OK(w, r, entry)
}

// GetEntry is the API handler to get a single entry.
func (c *Controller) GetEntry(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	builder := c.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if entry == nil {
		json.NotFound(w, r)
		return
	}

	json.OK(w, r, entry)
}

// GetFeedEntries is the API handler to get all feed entries.
func (c *Controller) GetFeedEntries(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")

	status := request.QueryStringParam(r, "status", "")
	if status != "" {
		if err := model.ValidateEntryStatus(status); err != nil {
			json.BadRequest(w, r, err)
			return
		}
	}

	order := request.QueryStringParam(r, "order", model.DefaultSortingOrder)
	if err := model.ValidateEntryOrder(order); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	direction := request.QueryStringParam(r, "direction", model.DefaultSortingDirection)
	if err := model.ValidateDirection(direction); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	limit := request.QueryIntParam(r, "limit", 100)
	offset := request.QueryIntParam(r, "offset", 0)
	if err := model.ValidateRange(offset, limit); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithFeedID(feedID)
	builder.WithStatus(status)
	builder.WithOrder(order)
	builder.WithDirection(direction)
	builder.WithOffset(offset)
	builder.WithLimit(limit)
	configureFilters(builder, r)

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, &entriesResponse{Total: count, Entries: entries})
}

// GetEntries is the API handler to fetch entries.
func (c *Controller) GetEntries(w http.ResponseWriter, r *http.Request) {
	status := request.QueryStringParam(r, "status", "")
	if status != "" {
		if err := model.ValidateEntryStatus(status); err != nil {
			json.BadRequest(w, r, err)
			return
		}
	}

	order := request.QueryStringParam(r, "order", model.DefaultSortingOrder)
	if err := model.ValidateEntryOrder(order); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	direction := request.QueryStringParam(r, "direction", model.DefaultSortingDirection)
	if err := model.ValidateDirection(direction); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	limit := request.QueryIntParam(r, "limit", 100)
	offset := request.QueryIntParam(r, "offset", 0)
	if err := model.ValidateRange(offset, limit); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithStatus(status)
	builder.WithOrder(order)
	builder.WithDirection(direction)
	builder.WithOffset(offset)
	builder.WithLimit(limit)
	configureFilters(builder, r)

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, &entriesResponse{Total: count, Entries: entries})
}

// SetEntryStatus is the API handler to change the status of entries.
func (c *Controller) SetEntryStatus(w http.ResponseWriter, r *http.Request) {
	entryIDs, status, err := decodeEntryStatusPayload(r.Body)
	if err != nil {
		json.BadRequest(w , r, errors.New("Invalid JSON payload"))
		return
	}

	if err := model.ValidateEntryStatus(status); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err := c.store.SetEntriesStatus(request.UserID(r), entryIDs, status); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

// ToggleBookmark is the API handler to toggle bookmark status.
func (c *Controller) ToggleBookmark(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	if err := c.store.ToggleBookmark(request.UserID(r), entryID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
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

	searchQuery := request.QueryStringParam(r, "search", "")
	if searchQuery != "" {
		builder.WithSearchQuery(searchQuery)
	}
}
