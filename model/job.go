// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

// Job represents a payload sent to the processing queue.
type Job struct {
	UserID int64
	FeedID int64
}

// JobList represents a list of jobs.
type JobList []Job
