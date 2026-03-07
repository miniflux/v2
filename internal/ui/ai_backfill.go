// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"
	"sync"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/integration/ai"
)

// aiBackfill triggers background AI summarization for unsummarized unread entries.
// If a backfill is already running for this user, it silently redirects without starting another.
func (h *handler) aiBackfill(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	// Skip if a backfill goroutine is already running for this user.
	if integration.IsBackfillRunning(userID) {
		html.Redirect(w, r, route.Path(h.router, "integrations"))
		return
	}

	userIntegrations, err := h.store.Integration(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Load user language for AI summary generation in the user's preferred language.
	user, err := h.store.UserByID(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	go integration.BackfillAISummaries(h.store, userID, userIntegrations, user.Language)

	html.Redirect(w, r, route.Path(h.router, "integrations"))
}

// aiForceBackfill triggers background AI re-summarization for ALL unread entries,
// overwriting existing summaries. Used when switching AI models or languages.
// If a backfill is already running for this user, it silently redirects without starting another.
func (h *handler) aiForceBackfill(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	// Skip if a backfill goroutine is already running for this user.
	if integration.IsBackfillRunning(userID) {
		html.Redirect(w, r, route.Path(h.router, "integrations"))
		return
	}

	userIntegrations, err := h.store.Integration(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Load user language for AI summary generation in the user's preferred language.
	user, err := h.store.UserByID(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	go integration.ForceBackfillAISummaries(h.store, userID, userIntegrations, user.Language)

	html.Redirect(w, r, route.Path(h.router, "integrations"))
}

// aiBackfillStatus returns JSON with the current backfill running state.
// Used by the integrations page JS to poll and update button states.
func (h *handler) aiBackfillStatus(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	json.OK(w, r, map[string]bool{"running": integration.IsBackfillRunning(userID)})
}

// aiStopBackfill signals the running backfill goroutine to stop.
func (h *handler) aiStopBackfill(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	integration.StopBackfill(userID)
	json.NoContent(w, r)
}

// pageSummaryResult stores the async result of a page summary generation task.
type pageSummaryResult struct {
	Status  string `json:"status"`            // "running", "done", "error"
	Summary string `json:"summary,omitempty"` // The generated digest text.
	Error   string `json:"error,omitempty"`   // Error message if generation failed.
}

// activePageSummaries stores in-progress and completed page summary results per user.
// Each user can only have one page summary task at a time.
var activePageSummaries sync.Map

// aiPageSummaryRequest holds entry IDs for generating a combined page summary.
type aiPageSummaryRequest struct {
	EntryIDs []int64 `json:"entry_ids"`
}

// aiPageSummary starts an async page summary generation task.
// Returns 202 Accepted immediately; the client polls aiPageSummaryStatus for the result.
func (h *handler) aiPageSummary(w http.ResponseWriter, r *http.Request) {
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

	// Collect individual AI summaries from entries to build a combined input.
	var summaryParts []string
	for _, entryID := range req.EntryIDs {
		builder := h.store.NewEntryQueryBuilder(userID)
		builder.WithEntryID(entryID)
		entry, entryErr := builder.GetEntry()
		if entryErr != nil || entry == nil {
			continue
		}
		if entry.AISummary != "" {
			summaryParts = append(summaryParts, entry.Title+": "+entry.AISummary)
		}
	}

	if len(summaryParts) == 0 {
		json.BadRequest(w, r, errors.New("no entries with AI summaries found"))
		return
	}

	combinedInput := ""
	for _, part := range summaryParts {
		combinedInput += part + "\n\n"
	}

	// Mark as running and launch async generation.
	activePageSummaries.Store(userID, &pageSummaryResult{Status: "running"})

	go func() {
		client := ai.NewClient(
			userIntegrations.AIProviderURL,
			userIntegrations.AIAPIKey,
			userIntegrations.AIModel,
		)

		result, err := client.GeneratePageSummary(combinedInput, user.Language)
		if err != nil {
			activePageSummaries.Store(userID, &pageSummaryResult{
				Status: "error",
				Error:  err.Error(),
			})
			return
		}

		activePageSummaries.Store(userID, &pageSummaryResult{
			Status:  "done",
			Summary: result,
		})
	}()

	json.Accepted(w, r)
}

// aiPageSummaryStatus returns the current status of the async page summary task.
// JS polls this endpoint until status is "done" or "error".
func (h *handler) aiPageSummaryStatus(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	val, ok := activePageSummaries.Load(userID)
	if !ok {
		json.OK(w, r, &pageSummaryResult{Status: "idle"})
		return
	}

	result := val.(*pageSummaryResult)
	json.OK(w, r, result)

	// Clean up completed/failed results after client retrieves them.
	// This avoids stale results showing up on future page loads.
	if result.Status == "done" || result.Status == "error" {
		activePageSummaries.Delete(userID)
	}
}
