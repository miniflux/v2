// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"miniflux.app/config"
	"miniflux.app/locale"
	"miniflux.app/reader/feed"
	"miniflux.app/scheduler"
	"miniflux.app/storage"
	"miniflux.app/template"

	"github.com/gorilla/mux"
)

// Controller contains all HTTP handlers for the user interface.
type Controller struct {
	cfg         *config.Config
	store       *storage.Storage
	pool        *scheduler.WorkerPool
	feedHandler *feed.Handler
	tpl         *template.Engine
	router      *mux.Router
	translator  *locale.Translator
}

// NewController returns a new Controller.
func NewController(cfg *config.Config, store *storage.Storage, pool *scheduler.WorkerPool, feedHandler *feed.Handler, tpl *template.Engine, translator *locale.Translator, router *mux.Router) *Controller {
	return &Controller{
		cfg:         cfg,
		store:       store,
		pool:        pool,
		feedHandler: feedHandler,
		tpl:         tpl,
		translator:  translator,
		router:      router,
	}
}
