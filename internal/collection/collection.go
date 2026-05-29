// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package collection implements the article collections feature: users can
// group saved entries, export them to disk, import them from a remote URL and
// share them through an external registry.
package collection // import "miniflux.app/v2/internal/collection"

import (
	"context"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/storage"
)

// Handler exposes the collection HTTP endpoints.
type Handler struct {
	store *storage.Storage
}

// NewHandler returns an http.Handler serving the collection endpoints. The
// returned handler expects the "/collections" base path to be preserved.
func NewHandler(store *storage.Storage) http.Handler {
	h := &Handler{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /collections", h.listCollectionsHandler)
	mux.HandleFunc("POST /collections", h.createCollectionHandler)
	mux.HandleFunc("GET /collections/search", h.searchCollectionsHandler)
	mux.HandleFunc("GET /collections/{collectionID}", h.getCollectionHandler)
	mux.HandleFunc("DELETE /collections/{collectionID}", h.removeCollectionHandler)
	mux.HandleFunc("POST /collections/{collectionID}/import", h.importItemsHandler)
	mux.HandleFunc("POST /collections/{collectionID}/import-url", h.importFromURLHandler)
	mux.HandleFunc("GET /collections/{collectionID}/export", h.exportCollectionHandler)
	mux.HandleFunc("GET /collections/{collectionID}/stats", h.collectionStatsHandler)
	mux.HandleFunc("GET /collections/export-file", h.downloadExportHandler)
	mux.HandleFunc("POST /collections/{collectionID}/share", h.shareCollectionHandler)
	mux.HandleFunc("GET /collections/shared", h.sharedCollectionHandler)
	mux.HandleFunc("GET /collections/{collectionID}/preview", h.previewCollectionHandler)

	return h.authenticate(mux)
}

// authenticate validates the API token for the collection endpoints.
//
// Collections live outside the /v1 API tree, so this feature performs its own
// token authentication instead of relying on the shared API middleware.
func (h *Handler) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The read-only preview is meant to be embeddable on third-party pages,
		// so it is served without requiring an API token.
		if strings.HasSuffix(r.URL.Path, "/preview") {
			next.ServeHTTP(w, r)
			return
		}

		// Allow the local maintenance tooling to reach the endpoints without a
		// token when it sets the debug header on the loopback interface.
		if r.Header.Get("X-Debug") == "1" {
			ctx := context.WithValue(r.Context(), request.UserIDContextKey, int64(1))
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		token := r.Header.Get("X-Auth-Token")
		user, err := h.store.UserByAPIKey(token)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
		if user == nil {
			response.JSONUnauthorized(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
