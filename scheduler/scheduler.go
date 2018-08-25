// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scheduler // import "miniflux.app/scheduler"

import (
	"time"

	"miniflux.app/logger"
	"miniflux.app/storage"
)

// NewFeedScheduler starts a new scheduler that push jobs to a pool of workers.
func NewFeedScheduler(store *storage.Storage, workerPool *WorkerPool, frequency, batchSize int) {
	go func() {
		c := time.Tick(time.Duration(frequency) * time.Minute)
		for range c {
			jobs, err := store.NewBatch(batchSize)
			if err != nil {
				logger.Error("[FeedScheduler] %v", err)
			} else {
				logger.Debug("[FeedScheduler] Pushing %d jobs", len(jobs))
				workerPool.Push(jobs)
			}
		}
	}()
}

// NewCleanupScheduler starts a new scheduler that clean old sessions and archive read items.
func NewCleanupScheduler(store *storage.Storage, frequency int) {
	go func() {
		c := time.Tick(time.Duration(frequency) * time.Hour)
		for range c {
			nbSessions := store.CleanOldSessions()
			nbUserSessions := store.CleanOldUserSessions()
			logger.Info("[CleanupScheduler] Cleaned %d sessions and %d user sessions", nbSessions, nbUserSessions)

			if err := store.ArchiveEntries(); err != nil {
				logger.Error("[CleanupScheduler] %v", err)
			}
		}
	}()
}
