// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/model"

// Job represents a payload sent to the processing queue.
type Job struct {
	UserID int64
	FeedID int64
}

// JobList represents a list of jobs.
type JobList []Job
