// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

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
	"miniflux.app/systemd"
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

	if systemd.HasNotifySocket() {
		logger.Info("Sending readiness notification to Systemd")

		if err := systemd.SdNotify(systemd.SdNotifyReady); err != nil {
			logger.Error("Unable to send readiness notification to systemd: %v", err)
		}

		if config.Opts.HasWatchdog() && systemd.HasSystemdWatchdog() {
			logger.Info("Activating Systemd watchdog")

			go func() {
				interval, err := systemd.WatchdogInterval()
				if err != nil {
					logger.Error("Unable to parse watchdog interval from systemd: %v", err)
					return
				}

				for {
					err := store.Ping()
					if err != nil {
						logger.Error(`Systemd Watchdog: %v`, err)
					} else {
						systemd.SdNotify(systemd.SdNotifyWatchdog)
					}

					time.Sleep(interval / 3)
				}
			}()
		}
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
