// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package worker // import "miniflux.app/worker"

import (
	"miniflux.app/model"
	"miniflux.app/storage"
)

// Pool handles a pool of workers.
type Pool struct {
	queue chan model.Job
}

// Push send a list of jobs to the queue.
func (p *Pool) Push(jobs model.JobList) {
	for _, job := range jobs {
		p.queue <- job
	}
}

// NewPool creates a pool of background workers.
func NewPool(store *storage.Storage, nbWorkers int) *Pool {
	workerPool := &Pool{
		queue: make(chan model.Job),
	}

	for i := 0; i < nbWorkers; i++ {
		worker := &Worker{id: i, store: store}
		go worker.Run(workerPool.queue)
	}

	return workerPool
}
