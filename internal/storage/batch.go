// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "influxeed-engine/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"strings"

	"influxeed-engine/v2/internal/model"
)

type BatchBuilder struct {
	db         *sql.DB
	args       []any
	conditions []string
	limit      int
}

func (s *Storage) NewBatchBuilder() *BatchBuilder {
	return &BatchBuilder{
		db: s.db,
	}
}

func (b *BatchBuilder) WithBatchSize(batchSize int) *BatchBuilder {
	b.limit = batchSize
	return b
}

func (b *BatchBuilder) WithUserID(userID int64) *BatchBuilder {
	b.conditions = append(b.conditions, fmt.Sprintf("user_id = $%d", len(b.args)+1))
	b.args = append(b.args, userID)
	return b
}

func (b *BatchBuilder) WithCategoryID(categoryID int64) *BatchBuilder {
	b.conditions = append(b.conditions, fmt.Sprintf("category_id = $%d", len(b.args)+1))
	b.args = append(b.args, categoryID)
	return b
}

func (b *BatchBuilder) WithErrorLimit(limit int) *BatchBuilder {
	if limit > 0 {
		b.conditions = append(b.conditions, fmt.Sprintf("parsing_error_count < $%d", len(b.args)+1))
		b.args = append(b.args, limit)
	}
	return b
}

func (b *BatchBuilder) WithNextCheckExpired() *BatchBuilder {
	b.conditions = append(b.conditions, "next_check_at < now()")
	return b
}

func (b *BatchBuilder) WithoutDisabledFeeds() *BatchBuilder {
	b.conditions = append(b.conditions, "disabled IS false")
	return b
}

func (b *BatchBuilder) FetchJobs() (model.JobList, error) {
	query := `SELECT id, user_id FROM feeds`

	if len(b.conditions) > 0 {
		query += " WHERE " + strings.Join(b.conditions, " AND ")
	}

	if b.limit > 0 {
		query += fmt.Sprintf(" ORDER BY next_check_at ASC LIMIT %d", b.limit)
	}

	rows, err := b.db.Query(query, b.args...)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch batch of jobs: %v`, err)
	}
	defer rows.Close()

	jobs := make(model.JobList, 0, b.limit)

	for rows.Next() {
		var job model.Job
		if err := rows.Scan(&job.FeedID, &job.UserID); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch job: %v`, err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
