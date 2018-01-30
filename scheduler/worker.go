// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scheduler

import (
	"time"

	"github.com/miniflux/miniflux/logger"
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
	logger.Info("[Worker] #%d started", w.id)

	for {
		job := <-c
		logger.Debug("[Worker #%d] got userID=%d, feedID=%d", w.id, job.UserID, job.FeedID)

		err := w.feedHandler.RefreshFeed(job.UserID, job.FeedID)
		if err != nil {
			logger.Error("[Worker] %v", err)
		}

		time.Sleep(time.Millisecond * 1000)
	}
}
