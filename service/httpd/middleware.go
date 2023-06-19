// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package httpd // import "miniflux.app/service/httpd"

import (
	"context"
	"net/http"

	"miniflux.app/config"
	"miniflux.app/http/request"
	"miniflux.app/logger"
)

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.FindClientIP(r)
		ctx := r.Context()
		ctx = context.WithValue(ctx, request.ClientIPContextKey, clientIP)

		if r.Header.Get("X-Forwarded-Proto") == "https" {
			config.Opts.HTTPS = true
		}

		protocol := "HTTP"
		if config.Opts.HTTPS {
			protocol = "HTTPS"
		}

		logger.Debug("[%s] %s %s %s", protocol, clientIP, r.Method, r.RequestURI)

		if config.Opts.HTTPS && config.Opts.HasHSTS() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
