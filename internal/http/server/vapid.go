// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server // import "miniflux.app/v2/internal/http/server"

import (
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/storage"
)

func newVAPIDProbe(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, publicKey, err := store.GetVAPIDKeys()

		if err != nil {
			http.Error(w, fmt.Sprintf("VAPID public key not found. Error: %q", err), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(publicKey))
	}
}
