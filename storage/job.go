// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"fmt"

	"miniflux.app/config"
	"miniflux.app/model"
)

// NewBatch returns a series of jobs.
func (s *Storage) NewBatch(batchSize int) (jobs model.JobList, err error) {
	pollingParsingErrorLimit := config.Opts.PollingParsingErrorLimit()
	query := `
		SELECT
			id,
			user_id
		FROM
			feeds
		WHERE
			disabled is false AND next_check_at < now() AND 
			CASE WHEN $1 > 0 THEN parsing_error_count < $1 ELSE parsing_error_count >= 0 END
		ORDER BY next_check_at ASC LIMIT $2
	`
	return s.fetchBatchRows(query, pollingParsingErrorLimit, batchSize)
}

// NewUserBatch returns a series of jobs but only for a given user.
func (s *Storage) NewUserBatch(userID int64, batchSize int) (jobs model.JobList, err error) {
	// We do not take the error counter into consideration when the given
	// user refresh manually all his feeds to force a refresh.
	query := `
		SELECT
			id,
			user_id
		FROM
			feeds
		WHERE
			user_id=$1 AND disabled is false
		ORDER BY next_check_at ASC LIMIT %d
	`
	return s.fetchBatchRows(fmt.Sprintf(query, batchSize), userID)
}

func (s *Storage) fetchBatchRows(query string, args ...interface{}) (jobs model.JobList, err error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch batch of jobs: %v`, err)
	}
	defer rows.Close()

	for rows.Next() {
		var job model.Job
		if err := rows.Scan(&job.FeedID, &job.UserID); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch job: %v`, err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
