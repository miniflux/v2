// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware // import "miniflux.app/middleware"

import (
	"context"
	"net/http"

	"miniflux.app/http/request"
)

// ClientIP stores in the real client IP address in the context.
func (m *Middleware) ClientIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, request.ClientIPContextKey, request.FindClientIP(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
