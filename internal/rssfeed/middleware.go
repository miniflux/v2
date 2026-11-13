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

// apiKeyParam is the query parameter that RSS readers can use to pass their
// Miniflux API key. Most feed readers support neither custom headers nor
// credentials, so the query parameter is the only universally supported way
// to authenticate a feed URL.
const apiKeyParam = "api_key"

// apiKeyAuthenticator resolves users from API keys and records key usage. It
// is an interface so that authentication can be tested without a database.
type apiKeyAuthenticator interface {
	UserByAPIKey(apiKey string) (*model.User, error)
	SetLastLogin(userID int64) error
	SetAPIKeyUsedTimestamp(userID int64, apiKey string) error
}

type authMiddleware struct {
	authenticator apiKeyAuthenticator
}

func newAuthMiddleware(authenticator apiKeyAuthenticator) *authMiddleware {
	return &authMiddleware{authenticator: authenticator}
}

// validateAPIKey authenticates the request with a Miniflux API key. The key
// can be provided as a query parameter (api_key), as an HTTP Basic password
// (with the Miniflux username), or as a bearer token.
func (m *authMiddleware) validateAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey, username := extractCredentials(r)

		if apiKey == "" {
			sendUnauthorized(w, r, "no API key provided")
			return
		}

		user, err := m.authenticator.UserByAPIKey(apiKey)
		if err != nil {
			slog.Error("[RSS] Unable to fetch user by API key",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", request.ClientIP(r)),
				slog.String("user_agent", r.UserAgent()),
				slog.Any("error", err),
			)
			sendUnauthorized(w, r, "invalid API key")
			return
		}

		if user == nil {
			slog.Warn("[RSS] No user found with the provided API key",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", request.ClientIP(r)),
				slog.String("user_agent", r.UserAgent()),
			)
			sendUnauthorized(w, r, "invalid API key")
			return
		}

		if username != "" && !strings.EqualFold(username, user.Username) {
			slog.Warn("[RSS] Username does not match the API key owner",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", request.ClientIP(r)),
				slog.String("user_agent", r.UserAgent()),
			)
			sendUnauthorized(w, r, "invalid credentials")
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
		m.authenticator.SetAPIKeyUsedTimestamp(user.ID, apiKey)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserNameContextKey, user.Username)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractCredentials returns the API key and the optional username sent with
// the request.
func extractCredentials(r *http.Request) (apiKey, username string) {
	if apiKey := r.URL.Query().Get(apiKeyParam); apiKey != "" {
		return apiKey, ""
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
	_, _ = w.Write([]byte("Unauthorized: provide a valid Miniflux API key as the " + apiKeyParam + " query parameter, as an HTTP Basic password (username = Miniflux username), or as a bearer token.\n"))
}
