// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"net/http"

	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/worker"

	"github.com/gorilla/mux"
)

type handler struct {
	store  *storage.Storage
	pool   *worker.Pool
	router *mux.Router
}

// Serve declares API routes for the application.
func Serve(router *mux.Router, store *storage.Storage, pool *worker.Pool) {
	handler := &handler{store, pool, router}

	sr := router.PathPrefix("/v1").Subrouter()
	middleware := newMiddleware(store)
	sr.Use(middleware.handleCORS)
	sr.Use(middleware.apiKeyAuth)
	sr.Use(middleware.basicAuth)
	sr.Methods(http.MethodOptions)
	sr.HandleFunc("/users", handler.createUserHandler).Methods(http.MethodPost)
	sr.HandleFunc("/users", handler.usersHandler).Methods(http.MethodGet)
	sr.HandleFunc("/users/{userID:[0-9]+}", handler.userByIDHandler).Methods(http.MethodGet)
	sr.HandleFunc("/users/{userID:[0-9]+}", handler.updateUserHandler).Methods(http.MethodPut)
	sr.HandleFunc("/users/{userID:[0-9]+}", handler.removeUserHandler).Methods(http.MethodDelete)
	sr.HandleFunc("/users/{userID:[0-9]+}/mark-all-as-read", handler.markUserAsReadHandler).Methods(http.MethodPut)
	sr.HandleFunc("/users/{username}", handler.userByUsernameHandler).Methods(http.MethodGet)
	sr.HandleFunc("/me", handler.currentUserHandler).Methods(http.MethodGet)
	sr.HandleFunc("/categories", handler.createCategoryHandler).Methods(http.MethodPost)
	sr.HandleFunc("/categories", handler.getCategoriesHandler).Methods(http.MethodGet)
	sr.HandleFunc("/categories/{categoryID}", handler.updateCategoryHandler).Methods(http.MethodPut)
	sr.HandleFunc("/categories/{categoryID}", handler.removeCategoryHandler).Methods(http.MethodDelete)
	sr.HandleFunc("/categories/{categoryID}/mark-all-as-read", handler.markCategoryAsReadHandler).Methods(http.MethodPut)
	sr.HandleFunc("/categories/{categoryID}/feeds", handler.getCategoryFeedsHandler).Methods(http.MethodGet)
	sr.HandleFunc("/categories/{categoryID}/refresh", handler.refreshCategoryHandler).Methods(http.MethodPut)
	sr.HandleFunc("/categories/{categoryID}/entries", handler.getCategoryEntriesHandler).Methods(http.MethodGet)
	sr.HandleFunc("/categories/{categoryID}/entries/{entryID}", handler.getCategoryEntryHandler).Methods(http.MethodGet)
	sr.HandleFunc("/discover", handler.discoverSubscriptionsHandler).Methods(http.MethodPost)
	sr.HandleFunc("/feeds", handler.createFeedHandler).Methods(http.MethodPost)
	sr.HandleFunc("/feeds", handler.getFeedsHandler).Methods(http.MethodGet)
	sr.HandleFunc("/feeds/counters", handler.fetchCountersHandler).Methods(http.MethodGet)
	sr.HandleFunc("/feeds/refresh", handler.refreshAllFeedsHandler).Methods(http.MethodPut)
	sr.HandleFunc("/feeds/{feedID}/refresh", handler.refreshFeedHandler).Methods(http.MethodPut)
	sr.HandleFunc("/feeds/{feedID}", handler.getFeedHandler).Methods(http.MethodGet)
	sr.HandleFunc("/feeds/{feedID}", handler.updateFeedHandler).Methods(http.MethodPut)
	sr.HandleFunc("/feeds/{feedID}", handler.removeFeedHandler).Methods(http.MethodDelete)
	sr.HandleFunc("/feeds/{feedID}/icon", handler.getIconByFeedIDHandler).Methods(http.MethodGet)
	sr.HandleFunc("/feeds/{feedID}/mark-all-as-read", handler.markFeedAsReadHandler).Methods(http.MethodPut)
	sr.HandleFunc("/export", handler.exportFeedsHandler).Methods(http.MethodGet)
	sr.HandleFunc("/import", handler.importFeedsHandler).Methods(http.MethodPost)
	sr.HandleFunc("/feeds/{feedID}/entries", handler.getFeedEntriesHandler).Methods(http.MethodGet)
	sr.HandleFunc("/feeds/{feedID}/entries/import", handler.importFeedEntryHandler).Methods(http.MethodPost)
	sr.HandleFunc("/feeds/{feedID}/entries/{entryID}", handler.getFeedEntryHandler).Methods(http.MethodGet)
	sr.HandleFunc("/entries", handler.getEntriesHandler).Methods(http.MethodGet)
	sr.HandleFunc("/entries", handler.setEntryStatusHandler).Methods(http.MethodPut)
	sr.HandleFunc("/entries/{entryID}", handler.getEntryHandler).Methods(http.MethodGet)
	sr.HandleFunc("/entries/{entryID}", handler.updateEntryHandler).Methods(http.MethodPut)
	sr.HandleFunc("/entries/{entryID}/bookmark", handler.toggleStarredHandler).Methods(http.MethodPut)
	sr.HandleFunc("/entries/{entryID}/star", handler.toggleStarredHandler).Methods(http.MethodPut)
	sr.HandleFunc("/entries/{entryID}/save", handler.saveEntryHandler).Methods(http.MethodPost)
	sr.HandleFunc("/entries/{entryID}/fetch-content", handler.fetchContentHandler).Methods(http.MethodGet)
	sr.HandleFunc("/flush-history", handler.flushHistoryHandler).Methods(http.MethodPut, http.MethodDelete)
	sr.HandleFunc("/icons/{iconID}", handler.getIconByIconIDHandler).Methods(http.MethodGet)
	sr.HandleFunc("/enclosures/{enclosureID}", handler.getEnclosureByIDHandler).Methods(http.MethodGet)
	sr.HandleFunc("/enclosures/{enclosureID}", handler.updateEnclosureByIDHandler).Methods(http.MethodPut)
	sr.HandleFunc("/integrations/status", handler.getIntegrationsStatusHandler).Methods(http.MethodGet)
	sr.HandleFunc("/version", handler.versionHandler).Methods(http.MethodGet)
	sr.HandleFunc("/api-keys", handler.createAPIKeyHandler).Methods(http.MethodPost)
	sr.HandleFunc("/api-keys", handler.getAPIKeysHandler).Methods(http.MethodGet)
	sr.HandleFunc("/api-keys/{apiKeyID}", handler.deleteAPIKeyHandler).Methods(http.MethodDelete)
}
