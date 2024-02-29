// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxy"
	"miniflux.app/v2/internal/reader/processor"
	"miniflux.app/v2/internal/reader/readingtime"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) getEntryFromBuilder(w http.ResponseWriter, r *http.Request, b *storage.EntryQueryBuilder) {
	entry, err := b.GetEntry()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if entry == nil {
		json.NotFound(w, r)
		return
	}

	entry.Content = proxy.AbsoluteProxyRewriter(h.router, r.Host, entry.Content)
	proxyOption := config.Opts.ProxyOption()

	for i := range entry.Enclosures {
		if proxyOption == "all" || proxyOption != "none" && !urllib.IsHTTPS(entry.Enclosures[i].URL) {
			for _, mediaType := range config.Opts.ProxyMediaTypes() {
				if strings.HasPrefix(entry.Enclosures[i].MimeType, mediaType+"/") {
					entry.Enclosures[i].URL = proxy.AbsoluteProxifyURL(h.router, r.Host, entry.Enclosures[i].URL)
					break
				}
			}
		}
	}

	json.OK(w, r, entry)
}

func (h *handler) getFeedEntry(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getCategoryEntry(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithCategoryID(categoryID)
	builder.WithEntryID(entryID)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getEntry(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithEntryID(entryID)

	h.getEntryFromBuilder(w, r, builder)
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
		json.BadRequest(w, r, errors.New("invalid category ID"))
		return
	}

	feedID = request.QueryInt64Param(r, "feed_id", feedID)
	if feedID > 0 && !h.store.FeedExists(userID, feedID) {
		json.BadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	tags := request.QueryStringParamList(r, "tags")

	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithFeedID(feedID)
	builder.WithCategoryID(categoryID)
	builder.WithStatuses(statuses)
	builder.WithSorting(order, direction)
	builder.WithOffset(offset)
	builder.WithLimit(limit)
	builder.WithTags(tags)
	builder.WithEnclosures()
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

	for i := range entries {
		entries[i].Content = proxy.AbsoluteProxyRewriter(h.router, r.Host, entries[i].Content)
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

func (h *handler) saveEntry(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	if !h.store.HasSaveEntry(request.UserID(r)) {
		json.BadRequest(w, r, errors.New("no third-party integration enabled"))
		return
	}

	entry, err := builder.GetEntry()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if entry == nil {
		json.NotFound(w, r)
		return
	}

	settings, err := h.store.Integration(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	go integration.SendEntry(entry, settings)

	json.Accepted(w, r)
}

func (h *handler) updateEntry(w http.ResponseWriter, r *http.Request) {
	var entryUpdateRequest model.EntryUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entryUpdateRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if err := validator.ValidateEntryModification(&entryUpdateRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

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

	user, err := h.store.UserByID(loggedUserID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	entryUpdateRequest.Patch(entry)
	if user.ShowReadingTime {
		entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
	}

	if err := h.store.UpdateEntryTitleAndContent(entry); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, entry)
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

	user, err := h.store.UserByID(loggedUserID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if user == nil {
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

	if err := processor.ProcessEntryWebPage(feed, entry, user); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, map[string]string{"content": entry.Content})
}

func (h *handler) flushHistory(w http.ResponseWriter, r *http.Request) {
	loggedUserID := request.UserID(r)
	go h.store.FlushHistory(loggedUserID)
	json.Accepted(w, r)
}

func configureFilters(builder *storage.EntryQueryBuilder, r *http.Request) {
	if beforeEntryID := request.QueryInt64Param(r, "before_entry_id", 0); beforeEntryID > 0 {
		builder.BeforeEntryID(beforeEntryID)
	}

	if afterEntryID := request.QueryInt64Param(r, "after_entry_id", 0); afterEntryID > 0 {
		builder.AfterEntryID(afterEntryID)
	}

	if beforePublishedTimestamp := request.QueryInt64Param(r, "before", 0); beforePublishedTimestamp > 0 {
		builder.BeforePublishedDate(time.Unix(beforePublishedTimestamp, 0))
	}

	if afterPublishedTimestamp := request.QueryInt64Param(r, "after", 0); afterPublishedTimestamp > 0 {
		builder.AfterPublishedDate(time.Unix(afterPublishedTimestamp, 0))
	}

	if beforePublishedTimestamp := request.QueryInt64Param(r, "published_before", 0); beforePublishedTimestamp > 0 {
		builder.BeforePublishedDate(time.Unix(beforePublishedTimestamp, 0))
	}

	if afterPublishedTimestamp := request.QueryInt64Param(r, "published_after", 0); afterPublishedTimestamp > 0 {
		builder.AfterPublishedDate(time.Unix(afterPublishedTimestamp, 0))
	}

	if beforeChangedTimestamp := request.QueryInt64Param(r, "changed_before", 0); beforeChangedTimestamp > 0 {
		builder.BeforeChangedDate(time.Unix(beforeChangedTimestamp, 0))
	}

	if afterChangedTimestamp := request.QueryInt64Param(r, "changed_after", 0); afterChangedTimestamp > 0 {
		builder.AfterChangedDate(time.Unix(afterChangedTimestamp, 0))
	}

	if categoryID := request.QueryInt64Param(r, "category_id", 0); categoryID > 0 {
		builder.WithCategoryID(categoryID)
	}

	if request.HasQueryParam(r, "starred") {
		starred, err := strconv.ParseBool(r.URL.Query().Get("starred"))
		if err == nil {
			builder.WithStarred(starred)
		}
	}

	if searchQuery := request.QueryStringParam(r, "search", ""); searchQuery != "" {
		builder.WithSearchQuery(searchQuery)
	}
}
