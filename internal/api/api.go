// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"net/http"

	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/worker"
)

type handler struct {
	store *storage.Storage
	pool  *worker.Pool
}

// NewHandler returns an http.Handler that handles API v1 calls.
// The returned handler expects the base path to be stripped from the request URL.
func NewHandler(store *storage.Storage, pool *worker.Pool) http.Handler {
	handler := &handler{store: store, pool: pool}
	middleware := newMiddleware(store)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/users", handler.createUserHandler)
	mux.HandleFunc("GET /v1/users", handler.usersHandler)
	mux.HandleFunc("GET /v1/users/{identifier}", handler.dispatchUserLookupHandler)
	mux.HandleFunc("PUT /v1/users/{userID}", handler.updateUserHandler)
	mux.HandleFunc("DELETE /v1/users/{userID}", handler.removeUserHandler)
	mux.HandleFunc("PUT /v1/users/{userID}/mark-all-as-read", handler.markUserAsReadHandler)
	mux.HandleFunc("GET /v1/me", handler.currentUserHandler)
	mux.HandleFunc("POST /v1/categories", handler.createCategoryHandler)
	mux.HandleFunc("GET /v1/categories", handler.getCategoriesHandler)
	mux.HandleFunc("PUT /v1/categories/{categoryID}", handler.updateCategoryHandler)
	mux.HandleFunc("DELETE /v1/categories/{categoryID}", handler.removeCategoryHandler)
	mux.HandleFunc("PUT /v1/categories/{categoryID}/mark-all-as-read", handler.markCategoryAsReadHandler)
	mux.HandleFunc("GET /v1/categories/{categoryID}/feeds", handler.getCategoryFeedsHandler)
	mux.HandleFunc("PUT /v1/categories/{categoryID}/refresh", handler.refreshCategoryHandler)
	mux.HandleFunc("GET /v1/categories/{categoryID}/entries", handler.getCategoryEntriesHandler)
	mux.HandleFunc("GET /v1/categories/{categoryID}/entries/{entryID}", handler.getCategoryEntryHandler)
	mux.HandleFunc("POST /v1/discover", handler.discoverSubscriptionsHandler)
	mux.HandleFunc("POST /v1/feeds", handler.createFeedHandler)
	mux.HandleFunc("GET /v1/feeds", handler.getFeedsHandler)
	mux.HandleFunc("GET /v1/feeds/counters", handler.fetchCountersHandler)
	mux.HandleFunc("PUT /v1/feeds/refresh", handler.refreshAllFeedsHandler)
	mux.HandleFunc("PUT /v1/feeds/{feedID}/refresh", handler.refreshFeedHandler)
	mux.HandleFunc("GET /v1/feeds/{feedID}", handler.getFeedHandler)
	mux.HandleFunc("PUT /v1/feeds/{feedID}", handler.updateFeedHandler)
	mux.HandleFunc("DELETE /v1/feeds/{feedID}", handler.removeFeedHandler)
	mux.HandleFunc("GET /v1/feeds/{feedID}/icon", handler.getIconByFeedIDHandler)
	mux.HandleFunc("PUT /v1/feeds/{feedID}/mark-all-as-read", handler.markFeedAsReadHandler)
	mux.HandleFunc("GET /v1/export", handler.exportFeedsHandler)
	mux.HandleFunc("POST /v1/import", handler.importFeedsHandler)
	mux.HandleFunc("GET /v1/feeds/{feedID}/entries", handler.getFeedEntriesHandler)
	mux.HandleFunc("POST /v1/feeds/{feedID}/entries/import", handler.importFeedEntryHandler)
	mux.HandleFunc("GET /v1/feeds/{feedID}/entries/{entryID}", handler.getFeedEntryHandler)
	mux.HandleFunc("GET /v1/entries", handler.getEntriesHandler)
	mux.HandleFunc("PUT /v1/entries", handler.setEntryStatusHandler)
	mux.HandleFunc("GET /v1/entries/{entryID}", handler.getEntryHandler)
	mux.HandleFunc("PUT /v1/entries/{entryID}", handler.updateEntryHandler)
	mux.HandleFunc("PUT /v1/entries/{entryID}/bookmark", handler.toggleStarredHandler)
	mux.HandleFunc("PUT /v1/entries/{entryID}/star", handler.toggleStarredHandler)
	mux.HandleFunc("POST /v1/entries/{entryID}/save", handler.saveEntryHandler)
	mux.HandleFunc("GET /v1/entries/{entryID}/fetch-content", handler.fetchContentHandler)
	mux.HandleFunc("PUT /v1/flush-history", handler.flushHistoryHandler)
	mux.HandleFunc("DELETE /v1/flush-history", handler.flushHistoryHandler)
	mux.HandleFunc("GET /v1/icons/{iconID}", handler.getIconByIconIDHandler)
	mux.HandleFunc("GET /v1/enclosures/{enclosureID}", handler.getEnclosureByIDHandler)
	mux.HandleFunc("PUT /v1/enclosures/{enclosureID}", handler.updateEnclosureByIDHandler)
	mux.HandleFunc("GET /v1/integrations/status", handler.getIntegrationsStatusHandler)
	mux.HandleFunc("GET /v1/version", handler.versionHandler)
	mux.HandleFunc("POST /v1/api-keys", handler.createAPIKeyHandler)
	mux.HandleFunc("GET /v1/api-keys", handler.getAPIKeysHandler)
	mux.HandleFunc("DELETE /v1/api-keys/{apiKeyID}", handler.deleteAPIKeyHandler)

	return middleware.withCORSHeaders(middleware.validateAPIKeyAuth(middleware.validateBasicAuth(mux)))
}
