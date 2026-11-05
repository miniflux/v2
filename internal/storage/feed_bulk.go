// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"fmt"
	"log/slog"

	"github.com/lib/pq"

	"miniflux.app/v2/internal/model"
)

// RemoveFeeds removes the given feeds along with their entries and enclosures.
// Feeds that do not belong to the user are ignored.
func (s *Storage) RemoveFeeds(userID int64, feedIDs []int64) error {
	if len(feedIDs) == 0 {
		return nil
	}

	query := `DELETE FROM feeds WHERE user_id=$1 AND id=ANY($2)`
	if _, err := s.db.Exec(query, userID, pq.Array(feedIDs)); err != nil {
		return fmt.Errorf(`store: unable to delete feeds: %v`, err)
	}
	return nil
}

// UpdateFeedsDisabled enables or disables the given feeds.
// Feeds that do not belong to the user are ignored.
func (s *Storage) UpdateFeedsDisabled(userID int64, feedIDs []int64, disabled bool) error {
	if len(feedIDs) == 0 {
		return nil
	}

	query := `UPDATE feeds SET disabled=$1 WHERE user_id=$2 AND id=ANY($3)`
	if _, err := s.db.Exec(query, disabled, userID, pq.Array(feedIDs)); err != nil {
		return fmt.Errorf(`store: unable to update feeds disabled flag: %v`, err)
	}
	return nil
}

// UpdateFeedsCategory moves the given feeds to another category.
// The category must belong to the same user. Feeds that do not belong to the user are ignored.
func (s *Storage) UpdateFeedsCategory(userID, categoryID int64, feedIDs []int64) error {
	if len(feedIDs) == 0 {
		return nil
	}

	query := `
		UPDATE
			feeds
		SET
			category_id=$1
		WHERE
			user_id=$2 AND id=ANY($3)
			AND EXISTS (SELECT 1 FROM categories WHERE id=$1 AND user_id=$2)
	`
	if _, err := s.db.Exec(query, categoryID, userID, pq.Array(feedIDs)); err != nil {
		return fmt.Errorf(`store: unable to move feeds to category #%d: %v`, categoryID, err)
	}
	return nil
}

// MarkFeedsAsRead marks as read the unread entries of the given feeds
// that were published before the last check of their feed.
// Feeds that do not belong to the user are ignored.
func (s *Storage) MarkFeedsAsRead(userID int64, feedIDs []int64) error {
	if len(feedIDs) == 0 {
		return nil
	}

	query := `
		UPDATE
			entries e
		SET
			status=$1,
			changed_at=now()
		FROM
			feeds f
		WHERE
			e.feed_id=f.id
			AND f.user_id=$2 AND f.id=ANY($3)
			AND e.user_id=$2 AND e.status=$4
			AND e.published_at < f.checked_at
	`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, pq.Array(feedIDs), model.EntryStatusUnread)
	if err != nil {
		return fmt.Errorf(`store: unable to mark feeds entries as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	slog.Debug("Marked feeds entries as read",
		slog.Int64("user_id", userID),
		slog.Any("feed_ids", feedIDs),
		slog.Int64("nb_entries", count),
	)

	return nil
}
