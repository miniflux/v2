// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package daemon // import "miniflux.app/daemon"

import (
	"miniflux.app/api"
	"miniflux.app/config"
	"miniflux.app/fever"
	"miniflux.app/middleware"
	"miniflux.app/reader/feed"
	"miniflux.app/scheduler"
	"miniflux.app/storage"
	"miniflux.app/ui"

	"github.com/gorilla/mux"
)

func routes(cfg *config.Config, store *storage.Storage, feedHandler *feed.Handler, pool *scheduler.WorkerPool) *mux.Router {
	router := mux.NewRouter()
	middleware := middleware.New(cfg, store, router)

	if cfg.BasePath() != "" {
		router = router.PathPrefix(cfg.BasePath()).Subrouter()
	}

	router.Use(middleware.ClientIP)
	router.Use(middleware.HeaderConfig)
	router.Use(middleware.Logging)

	fever.Serve(router, cfg, store)
	api.Serve(router, store, feedHandler)
	ui.Serve(router, cfg, store, pool, feedHandler)

	return router
}
