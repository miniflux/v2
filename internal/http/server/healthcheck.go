// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server // import "miniflux.app/v2/internal/http/server"

import (
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/storage"
)

func livenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func newReadinessProbe(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(); err != nil {
			http.Error(w, fmt.Sprintf("Database Connection Error: %q", err), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
