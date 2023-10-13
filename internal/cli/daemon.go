// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"miniflux.app/v2/internal/config"
	httpd "miniflux.app/v2/internal/http/server"
	"miniflux.app/v2/internal/metric"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/systemd"
	"miniflux.app/v2/internal/worker"
)

func startDaemon(store *storage.Storage) {
	slog.Debug("Starting daemon...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	pool := worker.NewPool(store, config.Opts.WorkerPoolSize())

	if config.Opts.HasSchedulerService() && !config.Opts.HasMaintenanceMode() {
		runScheduler(store, pool)
	}

	var httpServer *http.Server
	if config.Opts.HasHTTPService() {
		httpServer = httpd.StartWebServer(store, pool)
	}

	if config.Opts.HasMetricsCollector() {
		collector := metric.NewCollector(store, config.Opts.MetricsRefreshInterval())
		go collector.GatherStorageMetrics()
	}

	if systemd.HasNotifySocket() {
		slog.Debug("Sending readiness notification to Systemd")

		if err := systemd.SdNotify(systemd.SdNotifyReady); err != nil {
			slog.Error("Unable to send readiness notification to systemd", slog.Any("error", err))
		}

		if config.Opts.HasWatchdog() && systemd.HasSystemdWatchdog() {
			slog.Debug("Activating Systemd watchdog")

			go func() {
				interval, err := systemd.WatchdogInterval()
				if err != nil {
					slog.Error("Unable to get watchdog interval from systemd", slog.Any("error", err))
					return
				}

				for {
					if err := store.Ping(); err != nil {
						slog.Error("Unable to ping database", slog.Any("error", err))
					} else {
						systemd.SdNotify(systemd.SdNotifyWatchdog)
					}

					time.Sleep(interval / 3)
				}
			}()
		}
	}

	<-stop
	slog.Debug("Shutting down the process")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if httpServer != nil {
		httpServer.Shutdown(ctx)
	}

	slog.Debug("Process gracefully stopped")
}
