// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"log/slog"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/worker"
)

func runScheduler(store *storage.Storage, pool *worker.Pool) {
	slog.Debug(`Starting background scheduler...`)

	go feedScheduler(
		store,
		pool,
		config.Opts.PollingFrequency(),
		config.Opts.BatchSize(),
		config.Opts.PollingParsingErrorLimit(),
		config.Opts.PollingLimitPerHost(),
	)

	go cleanupScheduler(
		store,
		config.Opts.CleanupFrequency(),
	)
}

func feedScheduler(store *storage.Storage, pool *worker.Pool, frequency time.Duration, batchSize, errorLimit, limitPerHost int) {
	for range time.Tick(frequency) {
		// Generate a batch of feeds for any user that has feeds to refresh.
		batchBuilder := store.NewBatchBuilder()
		batchBuilder.WithBatchSize(batchSize)
		batchBuilder.WithErrorLimit(errorLimit)
		batchBuilder.WithoutDisabledFeeds()
		batchBuilder.WithNextCheckExpired()
		batchBuilder.WithLimitPerHost(limitPerHost)

		if jobs, err := batchBuilder.FetchJobs(); err != nil {
			slog.Error("Unable to fetch jobs from database", slog.Any("error", err))
		} else if len(jobs) > 0 {
			slog.Debug("Feed URLs in this batch", slog.Any("feed_urls", jobs.FeedURLs()))
			pool.Push(jobs)
		}
	}
}

func cleanupScheduler(store *storage.Storage, frequency time.Duration) {
	for range time.Tick(frequency) {
		runCleanupTasks(store)
	}
}
