// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"fmt"
	"log/slog"
	"time"

	"miniflux.app/v2/internal/model"
)

// TagExists checks if the given tag exists for the user.
func (s *Storage) TagExists(userID int64, tagName string) bool {
	var result bool
	query := `SELECT true FROM entries WHERE user_id=$1 AND LOWER($2) = ANY(LOWER(tags::text)::text[]) LIMIT 1`
	s.db.QueryRow(query, userID, tagName).Scan(&result)
	return result
}

// TagsWithEntryCount returns all entry tags with entry counts, sorted according to sortOrder.
func (s *Storage) TagsWithEntryCount(userID int64, sortOrder string) (model.Tags, error) {
	query := `
		WITH entry_tags AS (
			SELECT DISTINCT
				e.id,
				e.status,
				tag.title
			FROM entries e
			CROSS JOIN LATERAL unnest(e.tags) AS tag(title)
			WHERE
				e.user_id = $1 AND tag.title <> ''
		)
		SELECT
			title,
			count(*) AS count,
			count(*) FILTER (WHERE status = $2) AS count_unread
		FROM entry_tags
		GROUP BY title
	`

	if sortOrder == "alphabetical" {
		query += `
			ORDER BY
				title ASC
		`
	} else {
		query += `
			ORDER BY
				count_unread DESC,
				title ASC
		`
	}

	rows, err := s.db.Query(query, userID, model.EntryStatusUnread)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch tags: %v`, err)
	}
	defer rows.Close()

	tags := make(model.Tags, 0)
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.Title, &tag.TotalEntries, &tag.TotalUnread); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch tag row: %v`, err)
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// MarkTagAsRead updates all tag entries to the read status.
func (s *Storage) MarkTagAsRead(userID int64, tagName string, before time.Time) error {
	query := `
		UPDATE
			entries
		SET
			status=$1,
			changed_at=now()
		WHERE
			user_id=$2
		AND
			status=$3
		AND
			published_at < $4
		AND
			LOWER($5) = ANY(LOWER(tags::text)::text[])
	`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, model.EntryStatusUnread, before, tagName)
	if err != nil {
		return fmt.Errorf(`store: unable to mark tag entries as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	slog.Debug("Marked tag entries as read",
		slog.Int64("user_id", userID),
		slog.String("tag", tagName),
		slog.Int64("nb_entries", count),
		slog.String("before", before.Format(time.RFC3339)),
	)

	return nil
}
