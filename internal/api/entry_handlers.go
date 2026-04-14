// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/mediaproxy"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/processor"
	"miniflux.app/v2/internal/reader/readingtime"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) getEntryFromBuilder(w http.ResponseWriter, r *http.Request, b *storage.EntryQueryBuilder) {
	entry, err := b.GetEntry()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if entry == nil {
		response.JSONNotFound(w, r)
		return
	}

	entry.Content = mediaproxy.RewriteDocumentWithAbsoluteProxyURL(entry.Content)
	entry.Enclosures.ProxifyEnclosureURL(config.Opts.MediaProxyMode(), config.Opts.MediaProxyResourceTypes())

	response.JSON(w, r, entry)
}

func (h *handler) getFeedEntryHandler(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if feedID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")
	if entryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid entry ID"))
		return
	}

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getCategoryEntryHandler(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	if categoryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid category ID"))
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")
	if entryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid entry ID"))
		return
	}

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithCategoryID(categoryID)
	builder.WithEntryID(entryID)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getEntryHandler(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	if entryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid entry ID"))
		return
	}

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithEntryID(entryID)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getFeedEntriesHandler(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if feedID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	h.findEntries(w, r, feedID, 0)
}

func (h *handler) getCategoryEntriesHandler(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	if categoryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid category ID"))
		return
	}
	h.findEntries(w, r, 0, categoryID)
}

func (h *handler) getEntriesHandler(w http.ResponseWriter, r *http.Request) {
	h.findEntries(w, r, 0, 0)
}

func (h *handler) findEntries(w http.ResponseWriter, r *http.Request, feedID int64, categoryID int64) {
	statuses := request.QueryStringParamList(r, "status")
	for _, status := range statuses {
		if err := validator.ValidateEntryStatus(status); err != nil {
			response.JSONBadRequest(w, r, err)
			return
		}
	}

	order := request.QueryStringParam(r, "order", model.DefaultSortingOrder)
	if err := validator.ValidateEntryOrder(order); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	direction := request.QueryStringParam(r, "direction", model.DefaultSortingDirection)
	if err := validator.ValidateDirection(direction); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	limit := request.QueryIntParam(r, "limit", 100)
	offset := request.QueryIntParam(r, "offset", 0)
	if err := validator.ValidateRange(offset, limit); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	userID := request.UserID(r)
	categoryID = request.QueryInt64Param(r, "category_id", categoryID)
	if categoryID > 0 && !h.store.CategoryIDExists(userID, categoryID) {
		response.JSONBadRequest(w, r, errors.New("invalid category ID"))
		return
	}

	feedID = request.QueryInt64Param(r, "feed_id", feedID)
	if feedID > 0 && !h.store.FeedExists(userID, feedID) {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
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

	if request.HasQueryParam(r, "globally_visible") {
		globallyVisible := request.QueryBoolParam(r, "globally_visible", true)

		if globallyVisible {
			builder.WithGloballyVisible()
		}
	}

	configureFilters(builder, r)

	entries, count, err := builder.GetEntriesWithCount()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	for i := range entries {
		entries[i].Content = mediaproxy.RewriteDocumentWithAbsoluteProxyURL(entries[i].Content)
	}

	response.JSON(w, r, &entriesResponse{Total: count, Entries: entries})
}

func (h *handler) setEntryStatusHandler(w http.ResponseWriter, r *http.Request) {
	var entriesStatusUpdateRequest model.EntriesStatusUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entriesStatusUpdateRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if err := validator.ValidateEntriesStatusUpdateRequest(&entriesStatusUpdateRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if err := h.store.SetEntriesStatus(request.UserID(r), entriesStatusUpdateRequest.EntryIDs, entriesStatusUpdateRequest.Status); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}

func (h *handler) toggleStarredHandler(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	if entryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid entry ID"))
		return
	}

	if err := h.store.ToggleStarred(request.UserID(r), entryID); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}

func (h *handler) saveEntryHandler(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	if entryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid entry ID"))
		return
	}

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithEntryID(entryID)

	if !h.store.HasSaveEntry(request.UserID(r)) {
		response.JSONBadRequest(w, r, errors.New("no third-party integration enabled"))
		return
	}

	entry, err := builder.GetEntry()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if entry == nil {
		response.JSONNotFound(w, r)
		return
	}

	settings, err := h.store.Integration(request.UserID(r))
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	go integration.SendEntry(entry, settings)

	response.JSONAccepted(w, r)
}

func (h *handler) updateEntryHandler(w http.ResponseWriter, r *http.Request) {
	var entryUpdateRequest model.EntryUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entryUpdateRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if err := validator.ValidateEntryModification(&entryUpdateRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")
	if entryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid entry ID"))
		return
	}

	loggedUserID := request.UserID(r)
	entryBuilder := h.store.NewEntryQueryBuilder(loggedUserID)
	entryBuilder.WithEntryID(entryID)

	entry, err := entryBuilder.GetEntry()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if entry == nil {
		response.JSONNotFound(w, r)
		return
	}

	user, err := h.store.UserByID(loggedUserID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if user == nil {
		response.JSONNotFound(w, r)
		return
	}

	if entryUpdateRequest.Content != nil {
		sanitizedContent := sanitizer.SanitizeHTML(entry.URL, *entryUpdateRequest.Content, &sanitizer.SanitizerOptions{OpenLinksInNewTab: user.OpenExternalLinksInNewTab})
		entryUpdateRequest.Content = &sanitizedContent
	}

	entryUpdateRequest.Patch(entry)
	if user.ShowReadingTime {
		entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
	}

	if err := h.store.UpdateEntryTitleAndContent(entry); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSONCreated(w, r, entry)
}

func (h *handler) importFeedEntryHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	feedID := request.RouteInt64Param(r, "feedID")
	if feedID <= 0 {
		response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	if !h.store.FeedExists(userID, feedID) {
		response.JSONBadRequest(w, r, errors.New("feed does not exist"))
		return
	}

	var importRequest entryImportRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&importRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if importRequest.URL == "" {
		response.JSONBadRequest(w, r, errors.New("url is required"))
		return
	}

	if importRequest.Status == "" {
		importRequest.Status = model.EntryStatusRead
	}

	if err := validator.ValidateEntryStatus(importRequest.Status); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	entry := model.NewEntry()
	entry.URL = importRequest.URL
	entry.CommentsURL = importRequest.CommentsURL
	entry.Author = importRequest.Author
	entry.Tags = importRequest.Tags

	if importRequest.PublishedAt > 0 {
		entry.Date = time.Unix(importRequest.PublishedAt, 0).UTC()
	} else {
		entry.Date = time.Now().UTC()
	}

	if importRequest.Title == "" {
		entry.Title = entry.URL
	} else {
		entry.Title = importRequest.Title
	}

	hashInput := importRequest.ExternalID
	if hashInput == "" {
		hashInput = importRequest.URL
	}
	entry.Hash = crypto.HashFromBytes([]byte(hashInput))

	user, err := h.store.UserByID(userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if user == nil {
		response.JSONNotFound(w, r)
		return
	}

	if importRequest.Content != "" {
		entry.Content = sanitizer.SanitizeHTML(entry.URL, importRequest.Content, &sanitizer.SanitizerOptions{OpenLinksInNewTab: user.OpenExternalLinksInNewTab})
	}

	if user.ShowReadingTime {
		entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
	}

	created, err := h.store.InsertEntryForFeed(userID, feedID, entry)
	if errors.Is(err, storage.ErrEntryTombstoned) {
		response.JSONBadRequest(w, r, err)
		return
	}
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if err := h.store.SetEntriesStatus(userID, []int64{entry.ID}, importRequest.Status); err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	entry.Status = importRequest.Status

	if importRequest.Starred {
		if err := h.store.SetEntriesStarredState(userID, []int64{entry.ID}, true); err != nil {
			response.JSONServerError(w, r, err)
			return
		}
		entry.Starred = true
	}

	if created {
		response.JSONCreated(w, r, entryIDResponse{ID: entry.ID})
	} else {
		response.JSON(w, r, entryIDResponse{ID: entry.ID})
	}
}

func (h *handler) fetchContentHandler(w http.ResponseWriter, r *http.Request) {
	loggedUserID := request.UserID(r)

	entryID := request.RouteInt64Param(r, "entryID")
	if entryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid entry ID"))
		return
	}

	entryBuilder := h.store.NewEntryQueryBuilder(loggedUserID)
	entryBuilder.WithEntryID(entryID)

	entry, err := entryBuilder.GetEntry()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if entry == nil {
		response.JSONNotFound(w, r)
		return
	}

	user, err := h.store.UserByID(loggedUserID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if user == nil {
		response.JSONNotFound(w, r)
		return
	}

	feedBuilder := storage.NewFeedQueryBuilder(h.store, loggedUserID)
	feedBuilder.WithFeedID(entry.FeedID)
	feed, err := feedBuilder.GetFeed()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if feed == nil {
		response.JSONNotFound(w, r)
		return
	}

	if err := processor.ProcessEntryWebPage(feed, entry, user); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	shouldUpdateContent := request.QueryBoolParam(r, "update_content", false)
	if shouldUpdateContent {
		if err := h.store.UpdateEntryTitleAndContent(entry); err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	}

	response.JSON(w, r, entryContentResponse{Content: mediaproxy.RewriteDocumentWithAbsoluteProxyURL(entry.Content), ReadingTime: entry.ReadingTime})
}

func (h *handler) flushHistoryHandler(w http.ResponseWriter, r *http.Request) {
	loggedUserID := request.UserID(r)
	go h.store.FlushHistory(loggedUserID)
	response.JSONAccepted(w, r)
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
