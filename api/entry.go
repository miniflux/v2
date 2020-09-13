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

func (h *handler) getFeedEntry(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
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

func (h *handler) getEntry(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
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

func (h *handler) getFeedEntries(w http.ResponseWriter, r *http.Request) {
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

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
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

func (h *handler) getEntries(w http.ResponseWriter, r *http.Request) {
	statuses := request.QueryStringParamList(r, "status")
	for _, status := range statuses {
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

	userID := request.UserID(r)
	categoryID := request.QueryInt64Param(r, "category_id", 0)
	if categoryID > 0 && !h.store.CategoryExists(userID, categoryID) {
		json.BadRequest(w, r, errors.New("Invalid category ID"))
		return
	}

	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithCategoryID(categoryID)
	builder.WithStatuses(statuses)
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

func (h *handler) setEntryStatus(w http.ResponseWriter, r *http.Request) {
	entryIDs, status, err := decodeEntryStatusPayload(r.Body)
	if err != nil {
		json.BadRequest(w, r, errors.New("Invalid JSON payload"))
		return
	}

	if err := model.ValidateEntryStatus(status); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err := h.store.SetEntriesStatus(request.UserID(r), entryIDs, status); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) toggleBookmark(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	if err := h.store.ToggleBookmark(request.UserID(r), entryID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func configureFilters(builder *storage.EntryQueryBuilder, r *http.Request) {
	beforeEntryID := request.QueryInt64Param(r, "before_entry_id", 0)
	if beforeEntryID > 0 {
		builder.BeforeEntryID(beforeEntryID)
	}

	afterEntryID := request.QueryInt64Param(r, "after_entry_id", 0)
	if afterEntryID > 0 {
		builder.AfterEntryID(afterEntryID)
	}

	beforeTimestamp := request.QueryInt64Param(r, "before", 0)
	if beforeTimestamp > 0 {
		builder.BeforeDate(time.Unix(beforeTimestamp, 0))
	}

	afterTimestamp := request.QueryInt64Param(r, "after", 0)
	if afterTimestamp > 0 {
		builder.AfterDate(time.Unix(afterTimestamp, 0))
	}

	categoryID := request.QueryInt64Param(r, "category_id", 0)
	if categoryID > 0 {
		builder.WithCategoryID(categoryID)
	}

	if request.HasQueryParam(r, "starred") {
		builder.WithStarred()
	}

	searchQuery := request.QueryStringParam(r, "search", "")
	if searchQuery != "" {
		builder.WithSearchQuery(searchQuery)
	}
}
