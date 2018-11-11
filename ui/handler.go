// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"miniflux.app/config"
	"miniflux.app/reader/feed"
	"miniflux.app/scheduler"
	"miniflux.app/storage"
	"miniflux.app/template"

	"github.com/gorilla/mux"
)

type handler struct {
	router      *mux.Router
	cfg         *config.Config
	store       *storage.Storage
	tpl         *template.Engine
	pool        *scheduler.WorkerPool
	feedHandler *feed.Handler
}
