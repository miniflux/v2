// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever // import "miniflux.app/v2/internal/fever"

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

func (m *middleware) serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.ClientIP(r)
		apiKey := r.FormValue("api_key")
		if apiKey == "" {
			slog.Warn("[Fever] No API key provided",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			json.OK(w, r, newAuthFailureResponse())
			return
		}

		user, err := m.store.UserByFeverToken(apiKey)
		if err != nil {
			slog.Error("[Fever] Unable to fetch user by API key",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
				slog.Any("error", err),
			)
			json.OK(w, r, newAuthFailureResponse())
			return
		}

		if user == nil {
			slog.Warn("[Fever] No user found with the API key provided",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			json.OK(w, r, newAuthFailureResponse())
			return
		}

		slog.Info("[Fever] User authenticated successfully",
			slog.Bool("authentication_successful", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("user_id", user.ID),
			slog.String("username", user.Username),
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
