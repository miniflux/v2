// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware // import "miniflux.app/middleware"

import (
	"context"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/logger"
)

// BasicAuth handles HTTP basic authentication.
func (m *Middleware) BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		remoteAddr := request.RealIP(r)

		username, password, authOK := r.BasicAuth()
		if !authOK {
			logger.Debug("[Middleware:BasicAuth] No authentication headers sent")
			json.Unauthorized(w)
			return
		}

		if err := m.store.CheckPassword(username, password); err != nil {
			logger.Error("[Middleware:BasicAuth] [Remote=%v] Invalid username or password: %s", remoteAddr, username)
			json.Unauthorized(w)
			return
		}

		user, err := m.store.UserByUsername(username)
		if err != nil {
			logger.Error("[Middleware:BasicAuth] %v", err)
			json.ServerError(w, err)
			return
		}

		if user == nil {
			logger.Error("[Middleware:BasicAuth] [Remote=%v] User not found: %s", remoteAddr, username)
			json.Unauthorized(w)
			return
		}

		logger.Info("[Middleware:BasicAuth] User authenticated: %s", username)
		m.store.SetLastLogin(user.ID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
