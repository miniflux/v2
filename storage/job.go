// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"time"

	"github.com/miniflux/miniflux/helper"
	"github.com/miniflux/miniflux/model"
)

const maxParsingError = 3

// NewBatch returns a serie of jobs.
func (s *Storage) NewBatch(batchSize int) (jobs model.JobList, err error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:GetJobs] batchSize=%d", batchSize))
	query := `
		SELECT
		id, user_id
		FROM feeds
		WHERE parsing_error_count < $1
		ORDER BY checked_at ASC LIMIT %d`

	rows, err := s.db.Query(fmt.Sprintf(query, batchSize), maxParsingError)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch batch of jobs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var job model.Job
		if err := rows.Scan(&job.FeedID, &job.UserID); err != nil {
			return nil, fmt.Errorf("unable to fetch job: %v", err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
