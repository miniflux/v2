// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"
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

	builder := h.store.NewEntryQueryBuilder(request.UserID(r)).
		WithFeedID(feedID).
		WithEntryIDs(entryID)

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

	builder := h.store.NewEntryQueryBuilder(request.UserID(r)).
		WithCategoryID(categoryID).
		WithEntryIDs(entryID)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getEntryHandler(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	if entryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid entry ID"))
		return
	}

	builder := h.store.NewEntryQueryBuilder(request.UserID(r)).
		WithEntryIDs(entryID)

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
	filters, err := parseEntryFilters(r, feedID, categoryID)
	if err != nil {
		response.JSONBadRequest(w, r, err)
		return
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

	if !h.validateEntryFilterScope(w, r, filters) {
		return
	}

	builder := h.store.NewEntryQueryBuilder(request.UserID(r)).
		WithSorting(order, direction).
		WithOffset(offset).
		WithLimit(limit).
		WithEnclosures()

	filters.apply(builder)
	// Keep the unique tiebreak after any search ranking expression.
	if order != "id" {
		builder.WithSorting("id", direction)
	}

	entries, count, err := builder.GetEntriesWithCount()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	for i := range entries {
		entries[i].Content = mediaproxy.RewriteDocumentWithAbsoluteProxyURL(entries[i].Content)
		entries[i].Enclosures.ProxifyEnclosureURL(config.Opts.MediaProxyMode(), config.Opts.MediaProxyResourceTypes())
	}

	response.JSON(w, r, &entriesResponse{Total: count, Entries: entries})
}

func (h *handler) setEntryStatusAndStarredHandler(w http.ResponseWriter, r *http.Request) {
	var entriesStatusUpdateRequest model.EntriesStatusUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&entriesStatusUpdateRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if err := validator.ValidateEntriesStatusAndStarredUpdateRequest(&entriesStatusUpdateRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if entriesStatusUpdateRequest.Status != "" {
		if err := h.store.SetEntriesStatus(request.UserID(r), entriesStatusUpdateRequest.EntryIDs, entriesStatusUpdateRequest.Status); err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	}

	if entriesStatusUpdateRequest.Starred != nil {
		if err := h.store.SetEntriesStarredState(request.UserID(r), entriesStatusUpdateRequest.EntryIDs, *entriesStatusUpdateRequest.Starred); err != nil {
			response.JSONServerError(w, r, err)
			return
		}
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

	if !h.store.HasSaveEntry(request.UserID(r)) {
		response.JSONBadRequest(w, r, errors.New("no third-party integration enabled"))
		return
	}

	entry, err := h.store.NewEntryQueryBuilder(request.UserID(r)).
		WithEntryIDs(entryID).
		GetEntry()
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

	entry, err := h.store.NewEntryQueryBuilder(loggedUserID).
		WithEntryIDs(entryID).
		GetEntry()
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

	exists, err := h.store.FeedExists(userID, feedID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	if !exists {
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

	entry, err := h.store.NewEntryQueryBuilder(loggedUserID).
		WithEntryIDs(entryID).
		GetEntry()
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

	feed, err := h.store.NewFeedQueryBuilder(loggedUserID).
		WithFeedID(entry.FeedID).
		GetFeed()
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

func (h *handler) getEntryIDsHandler(w http.ResponseWriter, r *http.Request) {
	if err := validateEntryIDsFilters(r); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}
	filters, err := parseEntryFilters(r, 0, 0)
	if err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}
	if !h.validateEntryFilterScope(w, r, filters) {
		return
	}

	order := request.QueryStringParam(r, "order", "id")
	direction := request.QueryStringParam(r, "direction", "desc")
	limit, offset := parseEntryIDsParams(r)
	builder := h.store.NewEntryQueryBuilder(request.UserID(r)).
		WithSorting(order, direction).
		WithLimitAndMaximum(limit, model.MaxEntryIDsLimit).
		WithOffset(offset)

	filters.apply(builder)
	// Keep the unique tiebreak after any search ranking expression.
	if order != "id" {
		builder.WithSorting("id", direction)
	}

	entryIDs, total, err := builder.GetEntryIDsWithCount()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if entryIDs == nil {
		entryIDs = []int64{}
	}

	response.JSON(w, r, entryIDsResponse{Total: total, EntryIDs: entryIDs})
}

func (h *handler) flushHistoryHandler(w http.ResponseWriter, r *http.Request) {
	loggedUserID := request.UserID(r)
	go h.store.FlushHistory(loggedUserID)
	response.JSONAccepted(w, r)
}

func parseEntryIDsParams(r *http.Request) (limit, offset int) {
	limit = request.QueryIntParam(r, "limit", model.MaxEntryIDsLimit)
	if limit <= 0 || limit > model.MaxEntryIDsLimit {
		limit = model.MaxEntryIDsLimit
	}
	offset = request.QueryIntParam(r, "offset", 0)
	return limit, offset
}
