// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/integration"
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
