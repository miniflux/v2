// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server // import "miniflux.app/v2/internal/http/server"

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
)

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteIP := request.FindRemoteIP(r)
		isTrustedProxyClientIP := request.IsTrustedIP(remoteIP, config.Opts.TrustedReverseProxyNetworks())
		clientIP := request.FindClientIP(r, isTrustedProxyClientIP)
		ctx := r.Context()
		ctx = context.WithValue(ctx, request.ClientIPContextKey, clientIP)

		if isTrustedProxyClientIP && r.Header.Get("X-Forwarded-Proto") == "https" {
			config.Opts.SetHTTPSValue(true)
		}

		t1 := time.Now()
		defer func() {
			slog.Debug("Incoming request",
				slog.String("client_ip", clientIP),
				slog.Group("request",
					slog.String("method", r.Method),
					slog.String("uri", r.RequestURI),
					slog.String("protocol", r.Proto),
					slog.Duration("execution_time", time.Since(t1)),
				),
			)
		}()

		if config.Opts.HTTPS() && config.Opts.HasHSTS() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
