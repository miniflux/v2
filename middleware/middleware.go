// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"github.com/gorilla/mux"
	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/storage"
)

// Middleware handles different middleware handlers.
type Middleware struct {
	cfg    *config.Config
	store  *storage.Storage
	router *mux.Router
}

// New returns a new middleware.
func New(cfg *config.Config, store *storage.Storage, router *mux.Router) *Middleware {
	return &Middleware{cfg, store, router}
}
