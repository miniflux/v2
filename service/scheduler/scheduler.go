// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scheduler // import "miniflux.app/service/scheduler"

import (
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/storage"
	"miniflux.app/worker"
)

// Serve starts the internal scheduler.
func Serve(cfg *config.Config, store *storage.Storage, pool *worker.Pool) {
	logger.Info(`Starting scheduler...`)
	go store.CreateMediasRunOnce()
	go feedScheduler(store, pool, cfg.PollingFrequency(), cfg.BatchSize())
	go cleanupScheduler(store, cfg.CleanupFrequency(), cfg.ArchiveReadDays())
	if cfg.HasCacheService() {
		go cacheScheduler(store, cfg.CacheFrequency())
	}
}

func feedScheduler(store *storage.Storage, pool *worker.Pool, frequency, batchSize int) {
	c := time.Tick(time.Duration(frequency) * time.Minute)
	for range c {
		jobs, err := store.NewBatch(batchSize)
		if err != nil {
			logger.Error("[Scheduler:Feed] %v", err)
		} else {
			logger.Debug("[Scheduler:Feed] Pushing %d jobs", len(jobs))
			pool.Push(jobs)
		}
	}
}

func cleanupScheduler(store *storage.Storage, frequency int, archiveDays int) {
	c := time.Tick(time.Duration(frequency) * time.Hour)
	for range c {
		nbSessions := store.CleanOldSessions()
		nbUserSessions := store.CleanOldUserSessions()
		logger.Info("[Scheduler:Cleanup] Cleaned %d sessions and %d user sessions", nbSessions, nbUserSessions)

		if err := store.ArchiveEntries(archiveDays); err != nil {
			logger.Error("[Scheduler:Cleanup] %v", err)
		}
		if err := store.CleanupMedias(); err != nil {
			logger.Error("[Scheduler:Cleanup] %v", err)
		}
		if err := store.CleanupCaches(); err != nil {
			logger.Error("[Scheduler:Cleanup] %v", err)
		}
	}
}

func cacheScheduler(store *storage.Storage, frequency int) {
	c := time.Tick(time.Duration(frequency) * time.Hour)
	for range c {
		if err := store.CacheMedias(30); err != nil {
			logger.Error("[Scheduler:Cache] %v", err)
		}
	}
}
