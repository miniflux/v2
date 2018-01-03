// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package daemon

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/reader/feed"
	"github.com/miniflux/miniflux/scheduler"
	"github.com/miniflux/miniflux/storage"
)

// Run starts the daemon.
func Run(cfg *config.Config, store *storage.Storage) {
	logger.Info("Starting Miniflux...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	feedHandler := feed.NewFeedHandler(store)
	pool := scheduler.NewWorkerPool(feedHandler, cfg.GetInt("WORKER_POOL_SIZE", config.DefaultWorkerPoolSize))
	server := newServer(cfg, store, pool, feedHandler)

	scheduler.NewFeedScheduler(
		store,
		pool,
		cfg.GetInt("POLLING_FREQUENCY", config.DefaultPollingFrequency),
		cfg.GetInt("BATCH_SIZE", config.DefaultBatchSize),
	)

	scheduler.NewSessionScheduler(store, config.DefaultSessionCleanupFrequency)

	<-stop
	logger.Info("Shutting down the server...")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	server.Shutdown(ctx)
	store.Close()
	logger.Info("Server gracefully stopped")
}
