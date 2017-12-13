// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scheduler

import (
	"log"
	"time"

	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/reader/feed"
)

// Worker refreshes a feed in the background.
type Worker struct {
	id          int
	feedHandler *feed.Handler
}

// Run wait for a job and refresh the given feed.
func (w *Worker) Run(c chan model.Job) {
	log.Printf("[Worker] #%d started\n", w.id)

	for {
		job := <-c
		log.Printf("[Worker #%d] got userID=%d, feedID=%d\n", w.id, job.UserID, job.FeedID)

		err := w.feedHandler.RefreshFeed(job.UserID, job.FeedID)
		if err != nil {
			log.Println("Worker:", err)
		}

		time.Sleep(time.Millisecond * 1000)
	}
}
