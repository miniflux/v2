// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import "miniflux.app/v2/internal/urllib"

// Job represents a payload sent to the processing queue.
type Job struct {
	UserID  int64
	FeedID  int64
	FeedURL string
}

// JobList represents a list of jobs.
type JobList []Job

// FeedURLs returns the feed URLs of the job list with credentials redacted, for logging.
func (jl *JobList) FeedURLs() []string {
	feedURLs := make([]string, len(*jl))
	for i, job := range *jl {
		feedURLs[i] = urllib.RedactedURL(job.FeedURL)
	}
	return feedURLs
}
