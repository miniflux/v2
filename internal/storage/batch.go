// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"
)

type batchBuilder struct {
	db           *sql.DB
	args         []any
	where        whereBuilder
	batchSize    int
	limitPerHost int
}

func (s *Storage) NewBatchBuilder() *batchBuilder {
	b := batchBuilder{
		db: s.db,
	}

	return &b
}

func (b *batchBuilder) WithBatchSize(batchSize int) *batchBuilder {
	b.batchSize = batchSize
	return b
}

func (b *batchBuilder) WithUserID(userID int64) *batchBuilder {
	b.where.and("user_id = $" + strconv.Itoa(len(b.args)+1))
	b.args = append(b.args, userID)
	return b
}

func (b *batchBuilder) WithCategoryID(categoryID int64) *batchBuilder {
	b.where.and("category_id = $" + strconv.Itoa(len(b.args)+1))
	b.args = append(b.args, categoryID)
	return b
}

func (b *batchBuilder) WithErrorLimit(limit int) *batchBuilder {
	if limit > 0 {
		b.where.and("parsing_error_count < $" + strconv.Itoa(len(b.args)+1))
		b.args = append(b.args, limit)
	}
	return b
}

func (b *batchBuilder) WithNextCheckExpired() *batchBuilder {
	b.where.and("next_check_at < now()")
	return b
}

func (b *batchBuilder) WithoutDisabledFeeds() *batchBuilder {
	b.where.and("disabled IS false")
	return b
}

func (b *batchBuilder) WithLimitPerHost(limit int) *batchBuilder {
	if limit > 0 {
		b.limitPerHost = limit
	}
	return b
}

// FetchJobs retrieves a batch of jobs based on the conditions set in the builder.
// When limitPerHost is set, it limits the number of jobs per feed hostname to prevent overwhelming a single host.
func (b *batchBuilder) FetchJobs() (model.JobList, error) {
	var qb strings.Builder

	qb.WriteString(`
		SELECT id, user_id, feed_url
		FROM feeds
	`)

	qb.WriteString(" " + b.where.String())

	qb.WriteString(` ORDER BY next_check_at ASC`)

	if b.batchSize > 0 {
		qb.WriteString(` LIMIT ` + strconv.Itoa(b.batchSize))
	}

	rows, err := b.db.Query(qb.String(), b.args...)
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
