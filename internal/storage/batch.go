// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"
)

type BatchBuilder struct {
	db           *sql.DB
	args         []any
	conditions   []string
	batchSize    int
	limitPerHost int
}

func (s *Storage) NewBatchBuilder() *BatchBuilder {
	return &BatchBuilder{
		db: s.db,
	}
}

func (b *BatchBuilder) WithBatchSize(batchSize int) *BatchBuilder {
	b.batchSize = batchSize
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

func (b *BatchBuilder) WithLimitPerHost(limit int) *BatchBuilder {
	if limit > 0 {
		b.limitPerHost = limit
	}
	return b
}

// FetchJobs retrieves a batch of jobs based on the conditions set in the builder.
// It ensures that each job is unique by feed URL to avoid making too many concurrent requests to the same website.
// When limitPerHost is set, it limits the number of jobs per feed hostname to prevent overwhelming a single host.
func (b *BatchBuilder) FetchJobs() (model.JobList, error) {
	query := `SELECT DISTINCT ON (feed_url) id, user_id, feed_url FROM feeds`

	if len(b.conditions) > 0 {
		query += " WHERE " + strings.Join(b.conditions, " AND ")
	}

	query += " ORDER BY feed_url, next_check_at ASC"

	if b.batchSize > 0 {
		query += fmt.Sprintf(" LIMIT %d", b.batchSize)
	}

	rows, err := b.db.Query(query, b.args...)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch batch of jobs: %v`, err)
	}
	defer rows.Close()

	jobs := make(model.JobList, 0, b.batchSize)
	hosts := make(map[string]int)
	nbRows := 0
	nbSkippedFeeds := 0

	for rows.Next() {
		var job model.Job
		if err := rows.Scan(&job.FeedID, &job.UserID, &job.FeedURL); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch job record: %v`, err)
		}

		nbRows++

		if b.limitPerHost > 0 {
			feedHostname := urllib.Domain(job.FeedURL)
			if hosts[feedHostname] >= b.limitPerHost {
				slog.Debug("Feed host limit reached for this batch",
					slog.String("feed_url", job.FeedURL),
					slog.String("feed_hostname", feedHostname),
					slog.Int("limit_per_host", b.limitPerHost),
					slog.Int("current", hosts[feedHostname]),
				)
				nbSkippedFeeds++
				continue
			}
			hosts[feedHostname]++
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`store: error iterating on job records: %v`, err)
	}

	slog.Info("Created a batch of feeds",
		slog.Int("batch_size", b.batchSize),
		slog.Int("rows_count", nbRows),
		slog.Int("skipped_feeds_count", nbSkippedFeeds),
		slog.Int("jobs_count", len(jobs)),
	)

	return jobs, nil
}
