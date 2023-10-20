// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"log/slog"
	"sync"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	feedHandler "miniflux.app/v2/internal/reader/handler"
	"miniflux.app/v2/internal/storage"
)

func refreshFeeds(store *storage.Storage) {
	var wg sync.WaitGroup

	startTime := time.Now()
	jobs, err := store.NewBatch(config.Opts.BatchSize())
	if err != nil {
		slog.Error("Unable to fetch jobs from database", slog.Any("error", err))
		return
	}

	nbJobs := len(jobs)

	slog.Info("Created a batch of feeds",
		slog.Int("nb_jobs", nbJobs),
		slog.Int("batch_size", config.Opts.BatchSize()),
	)

	var jobQueue = make(chan model.Job, nbJobs)

	slog.Info("Starting a pool of workers",
		slog.Int("nb_workers", config.Opts.WorkerPoolSize()),
	)

	for i := 0; i < config.Opts.WorkerPoolSize(); i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobQueue {
				slog.Info("Refreshing feed",
					slog.Int64("feed_id", job.FeedID),
					slog.Int64("user_id", job.UserID),
					slog.Int("worker_id", workerID),
				)

				if err := feedHandler.RefreshFeed(store, job.UserID, job.FeedID, false); err != nil {
					slog.Warn("Unable to refresh feed",
						slog.Int64("feed_id", job.FeedID),
						slog.Int64("user_id", job.UserID),
						slog.Any("error", err),
					)
				}
			}
		}(i)
	}

	for _, job := range jobs {
		jobQueue <- job
	}
	close(jobQueue)

	wg.Wait()

	slog.Info("Refreshed a batch of feeds",
		slog.Int("nb_feeds", nbJobs),
		slog.String("duration", time.Since(startTime).String()),
	)
}
