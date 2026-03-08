// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/integration/ai"
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
		json.ServerError(w, r, err)
		return
	}

	if entry == nil {
		json.NotFound(w, r)
		return
	}

	entry.Content = mediaproxy.RewriteDocumentWithAbsoluteProxyURL(h.router, entry.Content)
	entry.Enclosures.ProxifyEnclosureURL(h.router, config.Opts.MediaProxyMode(), config.Opts.MediaProxyResourceTypes())

	json.OK(w, r, entry)
}

func (h *handler) getFeedEntry(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getCategoryEntry(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	entryID := request.RouteInt64Param(r, "entryID")

	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithCategoryID(categoryID)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	h.getEntryFromBuilder(w, r, builder)
}

func (h *handler) getEntry(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(request.UserID(r))
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

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
	builder.WithoutStatus(model.EntryStatusRemoved)

	if request.HasQueryParam(r, "globally_visible") {
		globallyVisible := request.QueryBoolParam(r, "globally_visible", true)

		if globallyVisible {
			builder.WithGloballyVisible()
		}
	}

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
		entries[i].Content = mediaproxy.RewriteDocumentWithAbsoluteProxyURL(h.router, entries[i].Content)
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

func (h *handler) toggleStarred(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	if err := h.store.ToggleStarred(request.UserID(r), entryID); err != nil {
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

	if entryUpdateRequest.Content != nil {
		sanitizedContent := sanitizer.SanitizeHTML(entry.URL, *entryUpdateRequest.Content, &sanitizer.SanitizerOptions{OpenLinksInNewTab: user.OpenExternalLinksInNewTab})
		entryUpdateRequest.Content = &sanitizedContent
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

func (h *handler) importFeedEntry(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	feedID := request.RouteInt64Param(r, "feedID")

	if feedID <= 0 {
		json.BadRequest(w, r, errors.New("invalid feed ID"))
		return
	}

	if !h.store.FeedExists(userID, feedID) {
		json.BadRequest(w, r, errors.New("feed does not exist"))
		return
	}

	var importRequest entryImportRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&importRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if importRequest.URL == "" {
		json.BadRequest(w, r, errors.New("url is required"))
		return
	}

	if importRequest.Status == "" {
		importRequest.Status = model.EntryStatusRead
	}

	if err := validator.ValidateEntryStatus(importRequest.Status); err != nil {
		json.BadRequest(w, r, err)
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
		json.ServerError(w, r, err)
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	if importRequest.Content != "" {
		entry.Content = sanitizer.SanitizeHTML(entry.URL, importRequest.Content, &sanitizer.SanitizerOptions{OpenLinksInNewTab: user.OpenExternalLinksInNewTab})
	}

	if user.ShowReadingTime {
		entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
	}

	created, err := h.store.InsertEntryForFeed(userID, feedID, entry)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if err := h.store.SetEntriesStatus(userID, []int64{entry.ID}, importRequest.Status); err != nil {
		json.ServerError(w, r, err)
		return
	}
	entry.Status = importRequest.Status

	if importRequest.Starred {
		if err := h.store.SetEntriesStarredState(userID, []int64{entry.ID}, true); err != nil {
			json.ServerError(w, r, err)
			return
		}
		entry.Starred = true
	}

	if created {
		json.Created(w, r, entryIDResponse{ID: entry.ID})
	} else {
		json.OK(w, r, entryIDResponse{ID: entry.ID})
	}
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

	shouldUpdateContent := request.QueryBoolParam(r, "update_content", false)
	if shouldUpdateContent {
		if err := h.store.UpdateEntryTitleAndContent(entry); err != nil {
			json.ServerError(w, r, err)
			return
		}
	}

	json.OK(w, r, entryContentResponse{Content: mediaproxy.RewriteDocumentWithAbsoluteProxyURL(h.router, entry.Content), ReadingTime: entry.ReadingTime})
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

func (h *handler) summarizeEntry(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	entryID := request.RouteInt64Param(r, "entryID")

	entryBuilder := h.store.NewEntryQueryBuilder(userID)
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

	userIntegrations, err := h.store.Integration(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if !userIntegrations.AIEnabled || userIntegrations.AIProviderURL == "" || userIntegrations.AIAPIKey == "" || userIntegrations.AIModel == "" {
		json.BadRequest(w, r, errors.New("AI integration is not configured"))
		return
	}

	client := ai.NewClient(
		userIntegrations.AIProviderURL,
		userIntegrations.AIAPIKey,
		userIntegrations.AIModel,
	)

	// Load user language for AI summary generation in the user's preferred language.
	user, err := h.store.UserByID(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	// Skip if already summarized — pass existing summary to let client decide
	result, err := client.SummarizeEntry(entry.Title, entry.Content, entry.AISummary, user.Language)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	// result is nil when entry already has a summary
	if result != nil {
		now := time.Now()
		entry.AISummary = result.Summary
		entry.AIScore = result.Score
		entry.AISummarizedAt = &now

		if err := h.store.UpdateEntryAISummary(entry); err != nil {
			json.ServerError(w, r, err)
			return
		}
	}

	json.OK(w, r, entry)
}

func (h *handler) backfillAISummaries(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	// Return 204 if a backfill is already running for this user — no duplicate work.
	if integration.IsBackfillRunning(userID) {
		json.NoContent(w, r)
		return
	}

	userIntegrations, err := h.store.Integration(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if !userIntegrations.AIEnabled || userIntegrations.AIProviderURL == "" || userIntegrations.AIAPIKey == "" || userIntegrations.AIModel == "" {
		json.BadRequest(w, r, errors.New("AI integration is not configured"))
		return
	}

	// Load user language for AI summary generation in the user's preferred language.
	user, err := h.store.UserByID(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	go integration.BackfillAISummaries(h.store, userID, userIntegrations, user.Language)

	json.Accepted(w, r)
}

func (h *handler) forceBackfillAISummaries(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	// Return 204 if a backfill is already running for this user — no duplicate work.
	if integration.IsBackfillRunning(userID) {
		json.NoContent(w, r)
		return
	}

	userIntegrations, err := h.store.Integration(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if !userIntegrations.AIEnabled || userIntegrations.AIProviderURL == "" || userIntegrations.AIAPIKey == "" || userIntegrations.AIModel == "" {
		json.BadRequest(w, r, errors.New("AI integration is not configured"))
		return
	}

	// Load user language for AI summary generation in the user's preferred language.
	user, err := h.store.UserByID(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	go integration.ForceBackfillAISummaries(h.store, userID, userIntegrations, user.Language)

	json.Accepted(w, r)
}

// aiPageSummaryRequest holds entry IDs for generating a combined page summary.
type aiPageSummaryRequest struct {
	EntryIDs []int64 `json:"entry_ids"`
}

// aiPageSummaryResponse returns the combined summary and the entry IDs that were summarized.
type aiPageSummaryResponse struct {
	Summary  string  `json:"summary"`
	EntryIDs []int64 `json:"entry_ids"`
}

// generateAIPageSummary takes a list of entry IDs, concatenates their AI summaries,
// and sends them to the AI provider for a combined digest summary.
func (h *handler) generateAIPageSummary(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	var req aiPageSummaryRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&req); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if len(req.EntryIDs) == 0 {
		json.BadRequest(w, r, errors.New("entry_ids is required"))
		return
	}

	userIntegrations, err := h.store.Integration(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if !userIntegrations.AIEnabled || userIntegrations.AIProviderURL == "" || userIntegrations.AIAPIKey == "" || userIntegrations.AIModel == "" {
		json.BadRequest(w, r, errors.New("AI integration is not configured"))
		return
	}

	// Load user language for AI summary generation in the user's preferred language.
	user, err := h.store.UserByID(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	// Collect individual AI summaries from entries to build a structured combined input.
	// Each entry is formatted with source (feed title) to help AI distinguish different authors/sources.
	var summaryParts []string
	for _, entryID := range req.EntryIDs {
		builder := h.store.NewEntryQueryBuilder(userID)
		builder.WithEntryID(entryID)
		entry, entryErr := builder.GetEntry()
		if entryErr != nil || entry == nil {
			continue
		}
		if entry.AISummary != "" {
			feedTitle := ""
			if entry.Feed != nil {
				feedTitle = entry.Feed.Title
			}
			summaryParts = append(summaryParts, fmt.Sprintf("[Source: %s] %s: %s", feedTitle, entry.Title, entry.AISummary))
		}
	}

	if len(summaryParts) == 0 {
		json.BadRequest(w, r, errors.New("no entries with AI summaries found"))
		return
	}

	// Build a combined input and send to AI for a digest summary.
	client := ai.NewClient(
		userIntegrations.AIProviderURL,
		userIntegrations.AIAPIKey,
		userIntegrations.AIModel,
	)

	combinedInput := ""
	for _, part := range summaryParts {
		combinedInput += part + "\n\n"
	}

	result, err := client.GeneratePageSummary(combinedInput, user.Language)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, &aiPageSummaryResponse{
		Summary:  result,
		EntryIDs: req.EntryIDs,
	})
}

func (h *handler) getBackfillStatus(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	json.OK(w, r, map[string]bool{"running": integration.IsBackfillRunning(userID)})
}

func (h *handler) stopBackfill(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	integration.StopBackfill(userID)
	json.NoContent(w, r)
}
