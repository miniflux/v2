// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server // import "miniflux.app/v2/internal/http/server"

import (
	"net/http"

	"miniflux.app/v2/internal/api"
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/fever"
	"miniflux.app/v2/internal/googlereader"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/ui"
	"miniflux.app/v2/internal/worker"
)

func newRouter(store *storage.Storage, pool *worker.Pool) http.Handler {
	readinessProbe := newReadinessProbe(store)

	// Application routes served under the base path.
	appMux := http.NewServeMux()

	appMux.HandleFunc("GET /healthcheck", readinessProbe)

	// Fever API routing.
	feverHandler := fever.Middleware(store)(fever.NewHandler(store))
	appMux.Handle("/fever/", feverHandler)

	// Google Reader API routing.
	googleReaderHandler := googlereader.NewHandler(store)
	appMux.HandleFunc("POST /accounts/ClientLogin", googleReaderHandler.ServeHTTP)
	appMux.Handle("/reader/api/0/", googleReaderHandler)

	// REST API routing.
	if config.Opts.HasAPI() {
		appMux.Handle("/v1/", api.NewHandler(store, pool))
	}

	// Metrics endpoint.
	if config.Opts.HasMetricsCollector() {
		appMux.Handle("GET /metrics", metricsHandler())
	}

	// UI routing (catch-all).
	appMux.Handle("/", ui.Serve(store, pool))

	// Apply shared middleware.
	var appHandler http.Handler = appMux
	if config.Opts.HasMaintenanceMode() {
		appHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(config.Opts.MaintenanceMessage()))
		})
	}
	appHandler = middleware(appHandler)

	// Root router: health probes at root, app routes under base path.
	rootMux := http.NewServeMux()

	// These routes do not take the base path into consideration and are always available at the root of the server.
	rootMux.HandleFunc("/liveness", livenessProbe)
	rootMux.HandleFunc("/healthz", livenessProbe)
	rootMux.HandleFunc("/readiness", readinessProbe)
	rootMux.HandleFunc("/readyz", readinessProbe)

	basePath := config.Opts.BasePath()
	if basePath != "" {
		rootMux.Handle(basePath+"/", http.StripPrefix(basePath, appHandler))
	} else {
		rootMux.Handle("/", appHandler)
	}

	return rootMux
}
