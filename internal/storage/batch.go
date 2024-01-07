// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"strings"

	"miniflux.app/v2/internal/model"
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
	b.conditions = append(b.conditions, "disabled is false")
	return b
}

func (b *BatchBuilder) FetchJobs() (jobs model.JobList, err error) {
	var parts []string
	parts = append(parts, `SELECT id, user_id FROM feeds`)

	if len(b.conditions) > 0 {
		parts = append(parts, fmt.Sprintf("WHERE %s", strings.Join(b.conditions, " AND ")))
	}

	if b.limit > 0 {
		parts = append(parts, fmt.Sprintf("ORDER BY next_check_at ASC LIMIT %d", b.limit))
	}

	query := strings.Join(parts, " ")
	rows, err := b.db.Query(query, b.args...)
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
