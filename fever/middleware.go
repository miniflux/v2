// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever // import "miniflux.app/fever"

import (
	"context"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/logger"
	"miniflux.app/storage"
)

type middleware struct {
	store *storage.Storage
}

func newMiddleware(s *storage.Storage) *middleware {
	return &middleware{s}
}

func (m *middleware) serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.ClientIP(r)
		apiKey := r.FormValue("api_key")
		if apiKey == "" {
			logger.Info("[Fever] [ClientIP=%s] No API key provided", clientIP)
			json.OK(w, r, newAuthFailureResponse())
			return
		}

		user, err := m.store.UserByFeverToken(apiKey)
		if err != nil {
			logger.Error("[Fever] %v", err)
			json.OK(w, r, newAuthFailureResponse())
			return
		}

		if user == nil {
			logger.Info("[Fever] [ClientIP=%s] No user found with this API key", clientIP)
			json.OK(w, r, newAuthFailureResponse())
			return
		}

		logger.Info("[Fever] [ClientIP=%s] User #%d is authenticated with user agent %q", clientIP, user.ID, r.UserAgent())
		m.store.SetLastLogin(user.ID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
