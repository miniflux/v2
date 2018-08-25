// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware // import "miniflux.app/middleware"

import (
	"net/http"
)

// HeaderConfig changes config values according to HTTP headers.
func (m *Middleware) HeaderConfig(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-Proto") == "https" {
			m.cfg.IsHTTPS = true
		}
		next.ServeHTTP(w, r)
	})
}
