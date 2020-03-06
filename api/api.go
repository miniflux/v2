// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"miniflux.app/reader/feed"
	"miniflux.app/storage"
	"miniflux.app/worker"

	"github.com/gorilla/mux"
)

// Serve declares API routes for the application.
func Serve(router *mux.Router, store *storage.Storage, pool *worker.Pool, feedHandler *feed.Handler) {
	handler := &handler{store, pool, feedHandler}

	sr := router.PathPrefix("/v1").Subrouter()
	middleware := newMiddleware(store)
	sr.Use(middleware.apiKeyAuth)
	sr.Use(middleware.basicAuth)
	sr.HandleFunc("/users", handler.createUser).Methods("POST")
	sr.HandleFunc("/users", handler.users).Methods("GET")
	sr.HandleFunc("/users/{userID:[0-9]+}", handler.userByID).Methods("GET")
	sr.HandleFunc("/users/{userID:[0-9]+}", handler.updateUser).Methods("PUT")
	sr.HandleFunc("/users/{userID:[0-9]+}", handler.removeUser).Methods("DELETE")
	sr.HandleFunc("/users/{username}", handler.userByUsername).Methods("GET")
	sr.HandleFunc("/me", handler.currentUser).Methods("GET")
	sr.HandleFunc("/categories", handler.createCategory).Methods("POST")
	sr.HandleFunc("/categories", handler.getCategories).Methods("GET")
	sr.HandleFunc("/categories/{categoryID}", handler.updateCategory).Methods("PUT")
	sr.HandleFunc("/categories/{categoryID}", handler.removeCategory).Methods("DELETE")
	sr.HandleFunc("/discover", handler.getSubscriptions).Methods("POST")
	sr.HandleFunc("/feeds", handler.createFeed).Methods("POST")
	sr.HandleFunc("/feeds", handler.getFeeds).Methods("GET")
	sr.HandleFunc("/feeds/refresh", handler.refreshAllFeeds).Methods("PUT")
	sr.HandleFunc("/feeds/{feedID}/refresh", handler.refreshFeed).Methods("PUT")
	sr.HandleFunc("/feeds/{feedID}", handler.getFeed).Methods("GET")
	sr.HandleFunc("/feeds/{feedID}", handler.updateFeed).Methods("PUT")
	sr.HandleFunc("/feeds/{feedID}", handler.removeFeed).Methods("DELETE")
	sr.HandleFunc("/feeds/{feedID}/icon", handler.feedIcon).Methods("GET")
	sr.HandleFunc("/export", handler.exportFeeds).Methods("GET")
	sr.HandleFunc("/import", handler.importFeeds).Methods("POST")
	sr.HandleFunc("/feeds/{feedID}/entries", handler.getFeedEntries).Methods("GET")
	sr.HandleFunc("/feeds/{feedID}/entries/{entryID}", handler.getFeedEntry).Methods("GET")
	sr.HandleFunc("/entries", handler.getEntries).Methods("GET")
	sr.HandleFunc("/entries", handler.setEntryStatus).Methods("PUT")
	sr.HandleFunc("/entries/{entryID}", handler.getEntry).Methods("GET")
	sr.HandleFunc("/entries/{entryID}/bookmark", handler.toggleBookmark).Methods("PUT")
}
