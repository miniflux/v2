// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scheduler

import (
	"github.com/miniflux/miniflux2/storage"
	"log"
	"time"
)

// NewScheduler starts a new scheduler to push jobs to a pool of workers.
func NewScheduler(store *storage.Storage, workerPool *WorkerPool, frequency, batchSize int) {
	c := time.Tick(time.Duration(frequency) * time.Minute)
	for now := range c {
		jobs := store.GetJobs(batchSize)
		log.Printf("[Scheduler:%v] => Pushing %d jobs\n", now, len(jobs))

		for _, job := range jobs {
			workerPool.Push(job)
		}
	}
}
