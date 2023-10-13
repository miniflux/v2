// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"context"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/storage"
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
			slog.Debug("[API] Skipped API token authentication because no API Key has been provided",
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.store.UserByAPIKey(token)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}

		if user == nil {
			slog.Warn("[API] No user found with the provided API key",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			json.Unauthorized(w, r)
			return
		}

		slog.Info("[API] User authenticated successfully with the API Token Authentication",
			slog.Bool("authentication_successful", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", user.Username),
		)

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
			slog.Warn("[API] No Basic HTTP Authentication header sent with the request",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			json.Unauthorized(w, r)
			return
		}

		if username == "" || password == "" {
			slog.Warn("[API] Empty username or password provided during Basic HTTP Authentication",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			json.Unauthorized(w, r)
			return
		}

		if err := m.store.CheckPassword(username, password); err != nil {
			slog.Warn("[API] Invalid username or password provided during Basic HTTP Authentication",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
				slog.String("username", username),
			)
			json.Unauthorized(w, r)
			return
		}

		user, err := m.store.UserByUsername(username)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}

		if user == nil {
			slog.Warn("[API] User not found while using Basic HTTP Authentication",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
				slog.String("username", username),
			)
			json.Unauthorized(w, r)
			return
		}

		slog.Info("[API] User authenticated successfully with the Basic HTTP Authentication",
			slog.Bool("authentication_successful", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", username),
		)

		m.store.SetLastLogin(user.ID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
