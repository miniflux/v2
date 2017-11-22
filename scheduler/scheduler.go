// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scheduler

import (
	"log"
	"time"

	"github.com/miniflux/miniflux2/storage"
)

// NewScheduler starts a new scheduler that push jobs to a pool of workers.
func NewScheduler(store *storage.Storage, workerPool *WorkerPool, frequency, batchSize int) {
	go func() {
		c := time.Tick(time.Duration(frequency) * time.Minute)
		for now := range c {
			jobs, err := store.NewBatch(batchSize)
			if err != nil {
				log.Println("[Scheduler]", err)
			} else {
				log.Printf("[Scheduler:%v] => Pushing %d jobs\n", now, len(jobs))
				workerPool.Push(jobs)
			}
		}
	}()
}
