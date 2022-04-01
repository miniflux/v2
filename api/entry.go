// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/model"
	"miniflux.app/reader/processor"
	"miniflux.app/storage"
	"miniflux.app/validator"
)

func getEntryFromBuilder(w http.ResponseWriter, r *http.Request, b *storage.EntryQueryBuilder) {
	entry, err := b.GetEntry()
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

func (h *handler) getFeedEntry(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	getEntryFromBuilder(w, r, builder)
}

func (h *handler) getCategoryEntry(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithCategoryID(categoryID)
	builder.WithEntryID(entryID)

	getEntryFromBuilder(w, r, builder)
}

func (h *handler) getEntry(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithEntryID(entryID)

	getEntryFromBuilder(w, r, builder)
}

func (h *handler) getFeedEntries(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	h.findEntries(w, r, feedID, 0)
}

func (h *handler) getCategoryEntries(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	h.findEntries(w, r, 0, categoryID)
}

func (h *handler) getEntries(w http.ResponseWriter, r *http.Request) {
	h.findEntries(w, r, 0, 0)
}

func (h *handler) findEntries(w http.ResponseWriter, r *http.Request, feedID int64, categoryID int64) {
	statuses := request.QueryStringParamList(r, "status")
	for _, status := range statuses {
		if err := validator.ValidateEntryStatus(status); err != nil {
			json.BadRequest(w, r, err)
			return
		}
	}

	order := request.QueryStringParam(r, "order", model.DefaultSortingOrder)
	if err := validator.ValidateEntryOrder(order); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	direction := request.QueryStringParam(r, "direction", model.DefaultSortingDirection)
	if err := validator.ValidateDirection(direction); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	limit := request.QueryIntParam(r, "limit", 100)
	offset := request.QueryIntParam(r, "offset", 0)
	if err := validator.ValidateRange(offset, limit); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	userID := request.UserID(r)
	categoryID = request.QueryInt64Param(r, "category_id", categoryID)
	if categoryID > 0 && !h.store.CategoryIDExists(userID, categoryID) {
		json.BadRequest(w, r, errors.New("Invalid category ID"))
		return
	}

	feedID = request.QueryInt64Param(r, "feed_id", feedID)
	if feedID > 0 && !h.store.FeedExists(userID, feedID) {
		json.BadRequest(w, r, errors.New("Invalid feed ID"))
		return
	}

	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithFeedID(feedID)
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
	var entriesStatusUpdateRequest model.EntriesStatusUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entriesStatusUpdateRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err := validator.ValidateEntriesStatusUpdateRequest(&entriesStatusUpdateRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err := h.store.SetEntriesStatus(request.UserID(r), entriesStatusUpdateRequest.EntryIDs, entriesStatusUpdateRequest.Status); err != nil {
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

func (h *handler) fetchContent(w http.ResponseWriter, r *http.Request) {
	loggedUserID := request.UserID(r)
	entryID := request.RouteInt64Param(r, "entryID")

	entryBuilder := h.store.NewEntryQueryBuilder(loggedUserID)
	entryBuilder.WithEntryID(entryID)
	entryBuilder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := entryBuilder.GetEntry()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if entry == nil {
		json.NotFound(w, r)
		return
	}

	feedBuilder := storage.NewFeedQueryBuilder(h.store, loggedUserID)
	feedBuilder.WithFeedID(entry.FeedID)
	feed, err := feedBuilder.GetFeed()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if feed == nil {
		json.NotFound(w, r)
		return
	}

	if err := processor.ProcessEntryWebPage(feed, entry); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, map[string]string{"content": entry.Content})
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
		starred, err := strconv.ParseBool(r.URL.Query().Get("starred"))
		if err == nil {
			builder.WithStarred(starred)
		}
	}

	searchQuery := request.QueryStringParam(r, "search", "")
	if searchQuery != "" {
		builder.WithSearchQuery(searchQuery)
	}
}
