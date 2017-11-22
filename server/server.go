// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package server

import (
	"log"
	"net/http"
	"time"

	"github.com/miniflux/miniflux2/config"
	"github.com/miniflux/miniflux2/reader/feed"
	"github.com/miniflux/miniflux2/storage"
)

// NewServer returns a new HTTP server.
func NewServer(cfg *config.Config, store *storage.Storage, feedHandler *feed.Handler) *http.Server {
	server := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Addr:         cfg.Get("LISTEN_ADDR", config.DefaultListenAddr),
		Handler:      getRoutes(cfg, store, feedHandler),
	}

	go func() {
		log.Printf("Listening on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	return server
}
