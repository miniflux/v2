// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package httpd // import "miniflux.app/service/httpd"

import (
	"context"
	"net/http"

	"miniflux.app/config"
	"miniflux.app/http/request"
	"miniflux.app/logger"
)

type middleware struct {
	cfg *config.Config
}

func newMiddleware(cfg *config.Config) *middleware {
	return &middleware{cfg}
}

func (m *middleware) Serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.FindClientIP(r)
		ctx := r.Context()
		ctx = context.WithValue(ctx, request.ClientIPContextKey, clientIP)

		if r.Header.Get("X-Forwarded-Proto") == "https" {
			m.cfg.IsHTTPS = true
		}

		protocol := "HTTP"
		if m.cfg.IsHTTPS {
			protocol = "HTTPS"
		}

		logger.Debug("[%s] %s %s %s", protocol, clientIP, r.Method, r.RequestURI)

		if m.cfg.IsHTTPS && m.cfg.HasHSTS() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
