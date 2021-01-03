// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli // import "miniflux.app/cli"

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/metric"
	"miniflux.app/service/httpd"
	"miniflux.app/service/scheduler"
	"miniflux.app/storage"
	"miniflux.app/worker"
)

func startDaemon(store *storage.Storage) {
	logger.Info("Starting Miniflux...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	pool := worker.NewPool(store, config.Opts.WorkerPoolSize())

	if config.Opts.HasSchedulerService() && !config.Opts.HasMaintenanceMode() {
		scheduler.Serve(store, pool)
	}

	var httpServer *http.Server
	if config.Opts.HasHTTPService() {
		httpServer = httpd.Serve(store, pool)
	}

	if config.Opts.HasMetricsCollector() {
		collector := metric.NewCollector(store, config.Opts.MetricsRefreshInterval())
		go collector.GatherStorageMetrics()
	}

	<-stop
	logger.Info("Shutting down the process...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if httpServer != nil {
		httpServer.Shutdown(ctx)
	}

	logger.Info("Process gracefully stopped")
}
