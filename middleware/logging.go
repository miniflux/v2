// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"

	"github.com/miniflux/miniflux/logger"
	"github.com/tomasen/realip"
)

// Logging logs the HTTP request.
func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[HTTP] %s %s %s", realip.RealIP(r), r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
