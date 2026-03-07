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

// aiBackfill triggers background AI summarization for all unsummarized entries.
// It launches a goroutine and immediately redirects back to the integrations page.
func (h *handler) aiBackfill(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	userIntegrations, err := h.store.Integration(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Run in background to avoid blocking the UI — summarization may take minutes.
	go integration.BackfillAISummaries(h.store, userID, userIntegrations)

	html.Redirect(w, r, route.Path(h.router, "integrations"))
}
