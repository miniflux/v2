// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/storage"
)

type authMiddleware struct {
	store *storage.Storage
}

func newAuthMiddleware(s *storage.Storage) *authMiddleware {
	return &authMiddleware{s}
}

func (m *authMiddleware) validateApiKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.serveValidated(w, r, next)
	})
}

func (m *authMiddleware) serveValidated(w http.ResponseWriter, r *http.Request, next http.Handler) {
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
			sendUnauthorizedResponse(w, r)
			return
		}

		token = r.Form.Get("T")
		if token == "" {
			slog.Warn("[GoogleReader] Post-Form T field is empty",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			sendUnauthorizedResponse(w, r)
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
			sendUnauthorizedResponse(w, r)
			return
		}
		fields := strings.Fields(authorization)
		if len(fields) != 2 {
			slog.Warn("[GoogleReader] Authorization header does not have the expected GoogleLogin format auth=xxxxxx",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			sendUnauthorizedResponse(w, r)
			return
		}
		if fields[0] != "GoogleLogin" {
			slog.Warn("[GoogleReader] Authorization header does not begin with GoogleLogin",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			sendUnauthorizedResponse(w, r)
			return
		}
		auths := strings.Split(fields[1], "=")
		if len(auths) != 2 {
			slog.Warn("[GoogleReader] Authorization header does not have the expected GoogleLogin format auth=xxxxxx",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			sendUnauthorizedResponse(w, r)
			return
		}
		if auths[0] != "auth" {
			slog.Warn("[GoogleReader] Authorization header does not have the expected GoogleLogin format auth=xxxxxx",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
			sendUnauthorizedResponse(w, r)
			return
		}
		token = auths[1]
	}

	parts := strings.Split(token, "/")
	if len(parts) != 2 || parts[0] == "" {
		slog.Warn("[GoogleReader] Auth token does not have the expected structure username/hash",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
		)
		sendUnauthorizedResponse(w, r)
		return
	}
	userID, err := m.store.GoogleReaderUserIDByToken(parts[0], parts[1])
	if err != nil {
		slog.Error("[GoogleReader] Unable to look up the token in the database",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Any("error", err),
		)
		sendUnauthorizedResponse(w, r)
		return
	}
	if userID == 0 {
		slog.Warn("[GoogleReader] No user found with the given token",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
		)
		sendUnauthorizedResponse(w, r)
		return
	}

	user, err := m.store.UserByID(userID)
	if err != nil {
		slog.Error("[GoogleReader] Unable to fetch user from database",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Any("error", err),
		)
		sendUnauthorizedResponse(w, r)
		return
	}

	if user == nil {
		slog.Warn("[GoogleReader] No user found with the given Google Reader credentials",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
		)
		sendUnauthorizedResponse(w, r)
		return
	}

	m.store.SetLastLogin(user.ID)

	ctx := r.Context()
	ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
	ctx = context.WithValue(ctx, request.UserNameContextKey, user.Username)
	ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
	ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
	ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)
	ctx = context.WithValue(ctx, request.GoogleReaderTokenKey, token)

	next.ServeHTTP(w, r.WithContext(ctx))
}
