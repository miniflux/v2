// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/model"
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
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *middleware) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.ClientIP(r)

		var token string
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				slog.Warn("[GoogleReader] Could not parse request form data",
					slog.Bool("authentication_failed", true),
					slog.String("client_ip", clientIP),
					slog.String("user_agent", r.UserAgent()),
					slog.Any("error", err),
				)
				Unauthorized(w, r)
				return
			}

			token = r.Form.Get("T")
			if token == "" {
				slog.Warn("[GoogleReader] Post-Form T field is empty",
					slog.Bool("authentication_failed", true),
					slog.String("client_ip", clientIP),
					slog.String("user_agent", r.UserAgent()),
				)
				Unauthorized(w, r)
				return
			}
		} else {
			authorization := r.Header.Get("Authorization")

			if authorization == "" {
				slog.Warn("[GoogleReader] No token provided",
					slog.Bool("authentication_failed", true),
					slog.String("client_ip", clientIP),
					slog.String("user_agent", r.UserAgent()),
				)
				Unauthorized(w, r)
				return
			}
			fields := strings.Fields(authorization)
			if len(fields) != 2 {
				slog.Warn("[GoogleReader] Authorization header does not have the expected GoogleLogin format auth=xxxxxx",
					slog.Bool("authentication_failed", true),
					slog.String("client_ip", clientIP),
					slog.String("user_agent", r.UserAgent()),
				)
				Unauthorized(w, r)
				return
			}
			if fields[0] != "GoogleLogin" {
				slog.Warn("[GoogleReader] Authorization header does not begin with GoogleLogin",
					slog.Bool("authentication_failed", true),
					slog.String("client_ip", clientIP),
					slog.String("user_agent", r.UserAgent()),
				)
				Unauthorized(w, r)
				return
			}
			auths := strings.Split(fields[1], "=")
			if len(auths) != 2 {
				slog.Warn("[GoogleReader] Authorization header does not have the expected GoogleLogin format auth=xxxxxx",
					slog.Bool("authentication_failed", true),
					slog.String("client_ip", clientIP),
					slog.String("user_agent", r.UserAgent()),
				)
				Unauthorized(w, r)
				return
			}
			if auths[0] != "auth" {
				slog.Warn("[GoogleReader] Authorization header does not have the expected GoogleLogin format auth=xxxxxx",
					slog.Bool("authentication_failed", true),
					slog.String("client_ip", clientIP),
					slog.String("user_agent", r.UserAgent()),
				)
				Unauthorized(w, r)
				return
			}
			token = auths[1]
		}

		parts := strings.Split(token, "/")
		if len(parts) != 2 {
			slog.Warn("[GoogleReader] Auth token does not have the expected structure username/hash",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
				slog.String("token", token),
			)
			Unauthorized(w, r)
			return
		}
		var integration *model.Integration
		var user *model.User
		var err error
		if integration, err = m.store.GoogleReaderUserGetIntegration(parts[0]); err != nil {
			slog.Warn("[GoogleReader] No user found with the given Google Reader username",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
				slog.Any("error", err),
			)
			Unauthorized(w, r)
			return
		}
		expectedToken := getAuthToken(integration.GoogleReaderUsername, integration.GoogleReaderPassword)
		if expectedToken != token {
			slog.Warn("[GoogleReader] Token does not match",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			Unauthorized(w, r)
			return
		}
		if user, err = m.store.UserByID(integration.UserID); err != nil {
			slog.Error("[GoogleReader] Unable to fetch user from database",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
				slog.Any("error", err),
			)
			Unauthorized(w, r)
			return
		}

		if user == nil {
			slog.Warn("[GoogleReader] No user found with the given Google Reader credentials",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			Unauthorized(w, r)
			return
		}

		slog.Info("[GoogleReader] User authenticated successfully",
			slog.Bool("authentication_successful", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("user_id", user.ID),
			slog.String("username", user.Username),
		)

		m.store.SetLastLogin(integration.UserID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)
		ctx = context.WithValue(ctx, request.GoogleReaderToken, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getAuthToken(username, password string) string {
	token := hex.EncodeToString(hmac.New(sha1.New, []byte(username+password)).Sum(nil))
	token = username + "/" + token
	return token
}
