// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssfeed // import "miniflux.app/v2/internal/rssfeed"

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/model"
)

// tokenParam is the query parameter that feed readers use to pass the feed
// token. Most feed readers support neither custom headers nor credentials, so
// the query parameter is the only universally supported way to authenticate a
// feed URL.
//
// The feed token is distinct from the REST API keys: it is generated from the
// Integrations page, only grants access to the RSS/Atom feeds, and is revoked
// by disabling the feeds. A leaked feed URL therefore never exposes the account.
const tokenParam = "token"

// tokenAuthenticator resolves users from feed tokens. It is an interface so
// that authentication can be tested without a database.
type tokenAuthenticator interface {
	UserByRSSFeedToken(token string) (*model.User, error)
	SetLastLogin(userID int64) error
}

type authMiddleware struct {
	authenticator tokenAuthenticator
}

func newAuthMiddleware(authenticator tokenAuthenticator) *authMiddleware {
	return &authMiddleware{authenticator: authenticator}
}

// validateToken authenticates the request with the user's feed token. The
// token can be provided as a query parameter (token), as an HTTP Basic
// password (with the Miniflux username), or as a bearer token.
func (m *authMiddleware) validateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, username := extractCredentials(r)

		if token == "" {
			sendUnauthorized(w, r, "no feed token provided")
			return
		}

		user, err := m.authenticator.UserByRSSFeedToken(token)
		if err != nil {
			slog.Error("[RSS] Unable to fetch user by feed token",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", request.ClientIP(r)),
				slog.String("user_agent", r.UserAgent()),
				slog.Any("error", err),
			)
			sendUnauthorized(w, r, "invalid feed token")
			return
		}

		if user == nil {
			sendUnauthorized(w, r, "invalid feed token")
			return
		}

		if username != "" && !strings.EqualFold(username, user.Username) {
			sendUnauthorized(w, r, "username does not match the feed token owner")
			return
		}

		slog.Info("[RSS] User authenticated successfully",
			slog.Bool("authentication_successful", true),
			slog.String("client_ip", request.ClientIP(r)),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("user_id", user.ID),
			slog.String("username", user.Username),
		)

		m.authenticator.SetLastLogin(user.ID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserNameContextKey, user.Username)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractCredentials returns the feed token and the optional username sent
// with the request.
func extractCredentials(r *http.Request) (token, username string) {
	if token := r.URL.Query().Get(tokenParam); token != "" {
		return token, ""
	}

	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return "", ""
	}

	if basicUsername, basicPassword, ok := r.BasicAuth(); ok {
		return basicPassword, basicUsername
	}

	scheme, token, found := strings.Cut(authorization, " ")
	if found && strings.EqualFold(scheme, "Bearer") {
		return strings.TrimSpace(token), ""
	}

	return "", ""
}

func sendUnauthorized(w http.ResponseWriter, r *http.Request, reason string) {
	slog.Warn("[RSS] Request rejected",
		slog.Bool("authentication_failed", true),
		slog.String("client_ip", request.ClientIP(r)),
		slog.String("user_agent", r.UserAgent()),
		slog.String("reason", reason),
	)

	w.Header().Set("WWW-Authenticate", `Basic realm="Miniflux RSS", charset="UTF-8"`)
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte("Unauthorized: enable the RSS/Atom feeds from the Integrations page and provide the feed token as the " + tokenParam + " query parameter, as an HTTP Basic password (username = Miniflux username), or as a bearer token.\n"))
}
