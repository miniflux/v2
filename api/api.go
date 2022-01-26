// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"net/http"

	"miniflux.app/storage"
	"miniflux.app/worker"

	"github.com/gorilla/mux"
)

type handler struct {
	store *storage.Storage
	pool  *worker.Pool
}

// Serve declares API routes for the application.
func Serve(router *mux.Router, store *storage.Storage, pool *worker.Pool) {
	handler := &handler{store, pool}

	sr := router.PathPrefix("/v1").Subrouter()
	middleware := newMiddleware(store)
	sr.Use(middleware.handleCORS)
	sr.Use(middleware.apiKeyAuth)
	sr.Use(middleware.basicAuth)
	sr.Methods(http.MethodOptions)
	sr.HandleFunc("/users", handler.createUser).Methods(http.MethodPost)
	sr.HandleFunc("/users", handler.users).Methods(http.MethodGet)
	sr.HandleFunc("/users/{userID:[0-9]+}", handler.userByID).Methods(http.MethodGet)
	sr.HandleFunc("/users/{userID:[0-9]+}", handler.updateUser).Methods(http.MethodPut)
	sr.HandleFunc("/users/{userID:[0-9]+}", handler.removeUser).Methods(http.MethodDelete)
	sr.HandleFunc("/users/{userID:[0-9]+}/mark-all-as-read", handler.markUserAsRead).Methods(http.MethodPut)
	sr.HandleFunc("/users/{username}", handler.userByUsername).Methods(http.MethodGet)
	sr.HandleFunc("/me", handler.currentUser).Methods(http.MethodGet)
	sr.HandleFunc("/categories", handler.createCategory).Methods(http.MethodPost)
	sr.HandleFunc("/categories", handler.getCategories).Methods(http.MethodGet)
	sr.HandleFunc("/categories/{categoryID}", handler.updateCategory).Methods(http.MethodPut)
	sr.HandleFunc("/categories/{categoryID}", handler.removeCategory).Methods(http.MethodDelete)
	sr.HandleFunc("/categories/{categoryID}/mark-all-as-read", handler.markCategoryAsRead).Methods(http.MethodPut)
	sr.HandleFunc("/categories/{categoryID}/feeds", handler.getCategoryFeeds).Methods(http.MethodGet)
	sr.HandleFunc("/categories/{categoryID}/entries", handler.getCategoryEntries).Methods(http.MethodGet)
	sr.HandleFunc("/categories/{categoryID}/entries/{entryID}", handler.getCategoryEntry).Methods(http.MethodGet)
	sr.HandleFunc("/discover", handler.discoverSubscriptions).Methods(http.MethodPost)
	sr.HandleFunc("/feeds", handler.createFeed).Methods(http.MethodPost)
	sr.HandleFunc("/feeds", handler.getFeeds).Methods(http.MethodGet)
	sr.HandleFunc("/feeds/refresh", handler.refreshAllFeeds).Methods(http.MethodPut)
	sr.HandleFunc("/feeds/{feedID}/refresh", handler.refreshFeed).Methods(http.MethodPut)
	sr.HandleFunc("/feeds/{feedID}", handler.getFeed).Methods(http.MethodGet)
	sr.HandleFunc("/feeds/{feedID}", handler.updateFeed).Methods(http.MethodPut)
	sr.HandleFunc("/feeds/{feedID}", handler.removeFeed).Methods(http.MethodDelete)
	sr.HandleFunc("/feeds/{feedID}/icon", handler.feedIcon).Methods(http.MethodGet)
	sr.HandleFunc("/feeds/{feedID}/mark-all-as-read", handler.markFeedAsRead).Methods(http.MethodPut)
	sr.HandleFunc("/export", handler.exportFeeds).Methods(http.MethodGet)
	sr.HandleFunc("/import", handler.importFeeds).Methods(http.MethodPost)
	sr.HandleFunc("/feeds/{feedID}/entries", handler.getFeedEntries).Methods(http.MethodGet)
	sr.HandleFunc("/feeds/{feedID}/entries/{entryID}", handler.getFeedEntry).Methods(http.MethodGet)
	sr.HandleFunc("/entries", handler.getEntries).Methods(http.MethodGet)
	sr.HandleFunc("/entries", handler.setEntryStatus).Methods(http.MethodPut)
	sr.HandleFunc("/entries/{entryID}", handler.getEntry).Methods(http.MethodGet)
	sr.HandleFunc("/entries/{entryID}/bookmark", handler.toggleBookmark).Methods(http.MethodPut)
	sr.HandleFunc("/entries/{entryID}/fetch-content", handler.fetchContent).Methods(http.MethodGet)
}
