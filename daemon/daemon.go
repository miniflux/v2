// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package daemon

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/locale"
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
	signal.Notify(stop, syscall.SIGTERM)

	go func() {
		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			logger.Debug("Alloc=%vK, TotalAlloc=%vK, Sys=%vK, NumGC=%v, GoRoutines=%d, NumCPU=%d",
				m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC, runtime.NumGoroutine(), runtime.NumCPU())
			time.Sleep(30 * time.Second)
		}
	}()

	translator := locale.Load()
	feedHandler := feed.NewFeedHandler(store, translator)
	pool := scheduler.NewWorkerPool(feedHandler, cfg.WorkerPoolSize())
	server := newServer(cfg, store, pool, feedHandler, translator)

	scheduler.NewFeedScheduler(
		store,
		pool,
		cfg.PollingFrequency(),
		cfg.BatchSize(),
	)

	scheduler.NewCleanupScheduler(store, cfg.CleanupFrequency())

	<-stop
	logger.Info("Shutting down the server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)
	store.Close()
	logger.Info("Server gracefully stopped")
}
