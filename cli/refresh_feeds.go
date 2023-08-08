// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/cli"

import (
	"sync"
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/model"
	feedHandler "miniflux.app/reader/handler"
	"miniflux.app/storage"
)

func refreshFeeds(store *storage.Storage) {
	var wg sync.WaitGroup

	startTime := time.Now()
	jobs, err := store.NewBatch(config.Opts.BatchSize())
	if err != nil {
		logger.Error("[Cronjob] %v", err)
	}

	nbJobs := len(jobs)
	logger.Info("[Cronjob]] Created %d jobs from a batch size of %d", nbJobs, config.Opts.BatchSize())
	var jobQueue = make(chan model.Job, nbJobs)

	logger.Info("[Cronjob] Starting a pool of %d workers", config.Opts.WorkerPoolSize())
	for i := 0; i < config.Opts.WorkerPoolSize(); i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobQueue {
				logger.Info("[Cronjob] Refreshing feed #%d for user #%d in worker #%d", job.FeedID, job.UserID, workerID)
				if err := feedHandler.RefreshFeed(store, job.UserID, job.FeedID, false); err != nil {
					logger.Error("[Cronjob] Refreshing the feed #%d returned this error: %v", job.FeedID, err)
				}
			}
		}(i)
	}

	for _, job := range jobs {
		jobQueue <- job
	}
	close(jobQueue)

	wg.Wait()
	logger.Info("[Cronjob] Refreshed %d feed(s) in %s", nbJobs, time.Since(startTime))
}
