// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package daemon // import "miniflux.app/daemon"

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/reader/feed"
	"miniflux.app/scheduler"
	"miniflux.app/storage"
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
			logger.Debug("Sys=%vK, InUse=%vK, HeapInUse=%vK, StackSys=%vK, StackInUse=%vK, GoRoutines=%d, NumCPU=%d",
				m.Sys/1024, (m.Sys-m.HeapReleased)/1024, m.HeapInuse/1024, m.StackSys/1024, m.StackInuse/1024,
				runtime.NumGoroutine(), runtime.NumCPU())
			time.Sleep(30 * time.Second)
		}
	}()

	feedHandler := feed.NewFeedHandler(store)
	pool := scheduler.NewWorkerPool(feedHandler, cfg.WorkerPoolSize())
	server := newServer(cfg, store, pool, feedHandler)

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
	logger.Info("Server gracefully stopped")
}
