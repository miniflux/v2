// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/cli"

import (
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/storage"
	"miniflux.app/worker"
)

func runScheduler(store *storage.Storage, pool *worker.Pool) {
	logger.Info(`Starting background scheduler...`)

	go feedScheduler(
		store,
		pool,
		config.Opts.PollingFrequency(),
		config.Opts.BatchSize(),
	)

	go cleanupScheduler(
		store,
		config.Opts.CleanupFrequencyHours(),
	)
}

func feedScheduler(store *storage.Storage, pool *worker.Pool, frequency, batchSize int) {
	for range time.Tick(time.Duration(frequency) * time.Minute) {
		jobs, err := store.NewBatch(batchSize)
		logger.Info("[Scheduler:Feed] Pushing %d jobs to the queue", len(jobs))
		if err != nil {
			logger.Error("[Scheduler:Feed] %v", err)
		} else {
			pool.Push(jobs)
		}
	}
}

func cleanupScheduler(store *storage.Storage, frequency int) {
	for range time.Tick(time.Duration(frequency) * time.Hour) {
		runCleanupTasks(store)
	}
}
