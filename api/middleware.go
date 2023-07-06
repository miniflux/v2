// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/api"

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
func (m *middleware) handleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "X-Auth-Token, Authorization, Content-Type, Accept")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *middleware) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.ClientIP(r)
		token := r.Header.Get("X-Auth-Token")

		if token == "" {
			logger.Debug("[API][TokenAuth] [ClientIP=%s] No API Key provided, go to the next middleware", clientIP)
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.store.UserByAPIKey(token)
		if err != nil {
			logger.Error("[API][TokenAuth] %v", err)
			json.ServerError(w, r, err)
			return
		}

		if user == nil {
			logger.Error("[API][TokenAuth] [ClientIP=%s] No user found with the given API key", clientIP)
			json.Unauthorized(w, r)
			return
		}

		logger.Info("[API][TokenAuth] [ClientIP=%s] User authenticated: %s", clientIP, user.Username)
		m.store.SetLastLogin(user.ID)
		m.store.SetAPIKeyUsedTimestamp(user.ID, token)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middleware) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if request.IsAuthenticated(r) {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		clientIP := request.ClientIP(r)
		username, password, authOK := r.BasicAuth()
		if !authOK {
			logger.Debug("[API][BasicAuth] [ClientIP=%s] No authentication headers sent", clientIP)
			json.Unauthorized(w, r)
			return
		}

		if username == "" || password == "" {
			logger.Error("[API][BasicAuth] [ClientIP=%s] Empty username or password", clientIP)
			json.Unauthorized(w, r)
			return
		}

		if err := m.store.CheckPassword(username, password); err != nil {
			logger.Error("[API][BasicAuth] [ClientIP=%s] Invalid username or password: %s", clientIP, username)
			json.Unauthorized(w, r)
			return
		}

		user, err := m.store.UserByUsername(username)
		if err != nil {
			logger.Error("[API][BasicAuth] %v", err)
			json.ServerError(w, r, err)
			return
		}

		if user == nil {
			logger.Error("[API][BasicAuth] [ClientIP=%s] User not found: %s", clientIP, username)
			json.Unauthorized(w, r)
			return
		}

		logger.Info("[API][BasicAuth] [ClientIP=%s] User authenticated: %s", clientIP, username)
		m.store.SetLastLogin(user.ID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
