// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package worker // import "miniflux.app/v2/internal/worker"

import (
	"log/slog"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/metric"
	"miniflux.app/v2/internal/model"
	feedHandler "miniflux.app/v2/internal/reader/handler"
	"miniflux.app/v2/internal/storage"
)

// worker refreshes a feed in the background.
type worker struct {
	id    int
	store *storage.Storage
}

// Run wait for a job and refresh the given feed.
func (w *worker) Run(c <-chan model.Job) {
	slog.Debug("Worker started",
		slog.Int("worker_id", w.id),
	)

	for {
		job := <-c
		slog.Debug("Job received by worker",
			slog.Int("worker_id", w.id),
			slog.Int64("user_id", job.UserID),
			slog.Int64("feed_id", job.FeedID),
			slog.String("feed_url", job.FeedURL),
		)

		startTime := time.Now()
		localizedError := feedHandler.RefreshFeed(w.store, job.UserID, job.FeedID, false)

		if config.Opts.HasMetricsCollector() {
			status := "success"
			if localizedError != nil {
				status = "error"
			}
			metric.BackgroundFeedRefreshDuration.WithLabelValues(status).Observe(time.Since(startTime).Seconds())
		}
	}
}
