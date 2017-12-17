// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scheduler

import (
	"time"

	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/storage"
)

// NewFeedScheduler starts a new scheduler that push jobs to a pool of workers.
func NewFeedScheduler(store *storage.Storage, workerPool *WorkerPool, frequency, batchSize int) {
	go func() {
		c := time.Tick(time.Duration(frequency) * time.Minute)
		for _ = range c {
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

// NewSessionScheduler starts a new scheduler that clean old sessions.
func NewSessionScheduler(store *storage.Storage, frequency int) {
	go func() {
		c := time.Tick(time.Duration(frequency) * time.Hour)
		for _ = range c {
			nbSessions := store.CleanOldSessions()
			nbUserSessions := store.CleanOldUserSessions()
			logger.Debug("[SessionScheduler] cleaned %d sessions and %d user sessions", nbSessions, nbUserSessions)
		}
	}()
}
