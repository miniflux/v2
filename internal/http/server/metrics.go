// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server // import "miniflux.app/v2/internal/http/server"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func metricsHandler() http.Handler {
	handler := promhttp.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAllowedToAccessMetricsEndpoint(r) {
			slog.Warn("Authentication failed while accessing the metrics endpoint",
				slog.String("client_ip", request.ClientIP(r)),
				slog.String("client_user_agent", r.UserAgent()),
				slog.String("client_remote_addr", r.RemoteAddr),
			)
			http.NotFound(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func isAllowedToAccessMetricsEndpoint(r *http.Request) bool {
	clientIP := request.ClientIP(r)

	if config.Opts.MetricsUsername() != "" && config.Opts.MetricsPassword() != "" {
		username, password, authOK := r.BasicAuth()
		if !authOK {
			slog.Warn("Metrics endpoint accessed without authentication header",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("client_user_agent", r.UserAgent()),
				slog.String("client_remote_addr", r.RemoteAddr),
			)
			return false
		}

		if username == "" || password == "" {
			slog.Warn("Metrics endpoint accessed with empty username or password",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("client_user_agent", r.UserAgent()),
				slog.String("client_remote_addr", r.RemoteAddr),
			)
			return false
		}

		// Both checks have to be run to avoid leaking informations
		// about the username and the password.
		usernameCorrect := crypto.ConstantTimeCmp(username, config.Opts.MetricsUsername())
		passwordCorrect := crypto.ConstantTimeCmp(password, config.Opts.MetricsPassword())
		if !usernameCorrect || !passwordCorrect {
			slog.Warn("Metrics endpoint accessed with invalid username or password",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("client_user_agent", r.UserAgent()),
				slog.String("client_remote_addr", r.RemoteAddr),
			)
			return false
		}
	}

	remoteIP := request.FindRemoteIP(r)
	return request.IsTrustedIP(remoteIP, config.Opts.MetricsAllowedNetworks())
}
