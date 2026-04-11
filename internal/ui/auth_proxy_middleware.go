// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

type authProxyMiddleware struct {
	basePath string
	store    *storage.Storage
}

func newAuthProxyMiddleware(basePath string, store *storage.Storage) *authProxyMiddleware {
	return &authProxyMiddleware{basePath: basePath, store: store}
}

func (m *authProxyMiddleware) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if request.IsAuthenticated(r) || config.Opts.AuthProxyHeader() == "" {
			next.ServeHTTP(w, r)
			return
		}

		remoteIP := request.FindRemoteIP(r)
		trustedNetworks := config.Opts.TrustedReverseProxyNetworks()
		if !request.IsTrustedIP(remoteIP, trustedNetworks) {
			slog.Warn("[AuthProxy] Rejecting authentication request from untrusted proxy",
				slog.String("remote_ip", remoteIP),
				slog.String("client_ip", request.ClientIP(r)),
				slog.String("user_agent", r.UserAgent()),
				slog.Any("trusted_networks", trustedNetworks),
			)
			next.ServeHTTP(w, r)
			return
		}

		username := r.Header.Get(config.Opts.AuthProxyHeader())
		if username == "" {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := request.ClientIP(r)
		slog.Debug("[AuthProxy] Received authenticated requested",
			slog.String("client_ip", clientIP),
			slog.String("remote_ip", remoteIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", username),
		)

		user, err := m.store.UserByUsername(username)
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		if user == nil {
			if !config.Opts.IsAuthProxyUserCreationAllowed() {
				slog.Debug("[AuthProxy] User doesn't exist and user creation is not allowed",
					slog.Bool("authentication_failed", true),
					slog.String("client_ip", clientIP),
					slog.String("remote_ip", remoteIP),
					slog.String("user_agent", r.UserAgent()),
					slog.String("username", username),
				)
				response.HTMLForbidden(w, r)
				return
			}

			if user, err = m.store.CreateUser(&model.UserCreationRequest{Username: username}); err != nil {
				response.HTMLServerError(w, r, err)
				return
			}
		}

		slog.Info("[AuthProxy] User authenticated successfully",
			slog.Bool("authentication_successful", true),
			slog.String("client_ip", clientIP),
			slog.String("remote_ip", remoteIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("user_id", user.ID),
			slog.String("username", user.Username),
		)

		m.store.SetLastLogin(user.ID)
		if err := authenticateWebSession(w, r, m.store, user); err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		response.HTMLRedirect(w, r, m.basePath+"/"+user.DefaultHomePage)
	})
}
