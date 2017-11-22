// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scheduler

import (
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/feed"
)

// WorkerPool handle a pool of workers.
type WorkerPool struct {
	queue chan model.Job
}

// Push send a list of jobs to the queue.
func (w *WorkerPool) Push(jobs model.JobList) {
	for _, job := range jobs {
		w.queue <- job
	}
}

// NewWorkerPool creates a pool of background workers.
func NewWorkerPool(feedHandler *feed.Handler, nbWorkers int) *WorkerPool {
	workerPool := &WorkerPool{
		queue: make(chan model.Job),
	}

	for i := 0; i < nbWorkers; i++ {
		worker := &Worker{id: i, feedHandler: feedHandler}
		go worker.Run(workerPool.queue)
	}

	return workerPool
}
