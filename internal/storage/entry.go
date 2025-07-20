// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"

	"github.com/lib/pq"
)

// CountAllEntries returns the number of entries for each status in the database.
func (s *Storage) CountAllEntries() map[string]int64 {
	rows, err := s.db.Query(`SELECT status, count(*) FROM entries GROUP BY status`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	results := make(map[string]int64)
	results[model.EntryStatusUnread] = 0
	results[model.EntryStatusRead] = 0
	results[model.EntryStatusRemoved] = 0

	for rows.Next() {
		var status string
		var count int64

		if err := rows.Scan(&status, &count); err != nil {
			continue
		}

		results[status] = count
	}

	results["total"] = results[model.EntryStatusUnread] + results[model.EntryStatusRead] + results[model.EntryStatusRemoved]
	return results
}

// CountUnreadEntries returns the number of unread entries.
func (s *Storage) CountUnreadEntries(userID int64) int {
	builder := s.NewEntryQueryBuilder(userID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithGloballyVisible()

	n, err := builder.CountEntries()
	if err != nil {
		slog.Error("Unable to count unread entries",
			slog.Int64("user_id", userID),
			slog.Any("error", err),
		)
		return 0
	}

	return n
}

// NewEntryQueryBuilder returns a new EntryQueryBuilder
func (s *Storage) NewEntryQueryBuilder(userID int64) *EntryQueryBuilder {
	return NewEntryQueryBuilder(s, userID)
}

// UpdateEntryTitleAndContent updates entry title and content.
func (s *Storage) UpdateEntryTitleAndContent(entry *model.Entry) error {
	query := `
		UPDATE
			entries
		SET
			title=$1,
			content=$2,
			reading_time=$3,
			document_vectors = setweight(to_tsvector($4), 'A') || setweight(to_tsvector($5), 'B')
		WHERE
			id=$6 AND user_id=$7
	`

	if _, err := s.db.Exec(
		query,
		entry.Title,
		entry.Content,
		entry.ReadingTime,
		truncateStringForTSVectorField(entry.Title),
		truncateStringForTSVectorField(entry.Content),
		entry.ID,
		entry.UserID); err != nil {
		return fmt.Errorf(`store: unable to update entry #%d: %v`, entry.ID, err)
	}

	return nil
}

// createEntry add a new entry.
func (s *Storage) createEntry(tx *sql.Tx, entry *model.Entry) error {
	query := `
		INSERT INTO entries
			(
				title,
				hash,
				url,
				comments_url,
				published_at,
				content,
				author,
				user_id,
				feed_id,
				reading_time,
				changed_at,
				document_vectors,
				tags
			)
		VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10,
				now(),
				setweight(to_tsvector($11), 'A') || setweight(to_tsvector($12), 'B'),
				$13
			)
		RETURNING
			id, status, created_at, changed_at
	`
	err := tx.QueryRow(
		query,
		entry.Title,
		entry.Hash,
		entry.URL,
		entry.CommentsURL,
		entry.Date,
		entry.Content,
		entry.Author,
		entry.UserID,
		entry.FeedID,
		entry.ReadingTime,
		truncateStringForTSVectorField(entry.Title),
		truncateStringForTSVectorField(entry.Content),
		pq.Array(entry.Tags),
	).Scan(
		&entry.ID,
		&entry.Status,
		&entry.CreatedAt,
		&entry.ChangedAt,
	)

	if err != nil {
		return fmt.Errorf(`store: unable to create entry %q (feed #%d): %v`, entry.URL, entry.FeedID, err)
	}

	for _, enclosure := range entry.Enclosures {
		enclosure.EntryID = entry.ID
		enclosure.UserID = entry.UserID
		err := s.createEnclosure(tx, enclosure)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateEntry updates an entry when a feed is refreshed.
// Note: we do not update the published date because some feeds do not contains any date,
// it default to time.Now() which could change the order of items on the history page.
func (s *Storage) updateEntry(tx *sql.Tx, entry *model.Entry) error {
	query := `
		UPDATE
			entries
		SET
			title=$1,
			url=$2,
			comments_url=$3,
			content=$4,
			author=$5,
			reading_time=$6,
			document_vectors = setweight(to_tsvector($7), 'A') || setweight(to_tsvector($8), 'B'),
			tags=$12
		WHERE
			user_id=$9 AND feed_id=$10 AND hash=$11
		RETURNING
			id
	`
	err := tx.QueryRow(
		query,
		entry.Title,
		entry.URL,
		entry.CommentsURL,
		entry.Content,
		entry.Author,
		entry.ReadingTime,
		truncateStringForTSVectorField(entry.Title),
		truncateStringForTSVectorField(entry.Content),
		entry.UserID,
		entry.FeedID,
		entry.Hash,
		pq.Array(entry.Tags),
	).Scan(&entry.ID)

	if err != nil {
		return fmt.Errorf(`store: unable to update entry %q: %v`, entry.URL, err)
	}

	for _, enclosure := range entry.Enclosures {
		enclosure.UserID = entry.UserID
		enclosure.EntryID = entry.ID
	}

	return s.updateEnclosures(tx, entry)
}

// entryExists checks if an entry already exists based on its hash when refreshing a feed.
func (s *Storage) entryExists(tx *sql.Tx, entry *model.Entry) (bool, error) {
	var result bool

	// Note: This query uses entries_feed_id_hash_key index (filtering on user_id is not necessary).
	err := tx.QueryRow(`SELECT true FROM entries WHERE feed_id=$1 AND hash=$2`, entry.FeedID, entry.Hash).Scan(&result)

	if err != nil && err != sql.ErrNoRows {
		return result, fmt.Errorf(`store: unable to check if entry exists: %v`, err)
	}

	return result, nil
}

func (s *Storage) IsNewEntry(feedID int64, entryHash string) bool {
	var result bool
	s.db.QueryRow(`SELECT true FROM entries WHERE feed_id=$1 AND hash=$2`, feedID, entryHash).Scan(&result)
	return !result
}

func (s *Storage) GetReadTime(feedID int64, entryHash string) int {
	var result int

	// Note: This query uses entries_feed_id_hash_key index
	s.db.QueryRow(
		`SELECT
			reading_time
		FROM
			entries
		WHERE
			feed_id=$1 AND
			hash=$2
		`,
		feedID,
		entryHash,
	).Scan(&result)
	return result
}

// cleanupEntries deletes from the database entries marked as "removed" and not visible anymore in the feed.
func (s *Storage) cleanupEntries(feedID int64, entryHashes []string) error {
	query := `
		DELETE FROM
			entries
		WHERE
			feed_id=$1 AND
			status=$2 AND
			NOT (hash=ANY($3))
	`
	if _, err := s.db.Exec(query, feedID, model.EntryStatusRemoved, pq.Array(entryHashes)); err != nil {
		return fmt.Errorf(`store: unable to cleanup entries: %v`, err)
	}

	return nil
}

// RefreshFeedEntries updates feed entries while refreshing a feed.
func (s *Storage) RefreshFeedEntries(userID, feedID int64, entries model.Entries, updateExistingEntries bool) (newEntries model.Entries, err error) {
	entryHashes := make([]string, 0, len(entries))

	for _, entry := range entries {
		entry.UserID = userID
		entry.FeedID = feedID

		tx, err := s.db.Begin()
		if err != nil {
			return nil, fmt.Errorf(`store: unable to start transaction: %v`, err)
		}

		entryExists, err := s.entryExists(tx, entry)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf(`store: unable to rollback transaction: %v (rolled back due to: %v)`, rollbackErr, err)
			}
			return nil, err
		}

		if entryExists {
			if updateExistingEntries {
				err = s.updateEntry(tx, entry)
			}
		} else {
			err = s.createEntry(tx, entry)
			if err == nil {
				newEntries = append(newEntries, entry)
			}
		}

		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf(`store: unable to rollback transaction: %v (rolled back due to: %v)`, rollbackErr, err)
			}
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf(`store: unable to commit transaction: %v`, err)
		}

		entryHashes = append(entryHashes, entry.Hash)
	}

	go func() {
		if err := s.cleanupEntries(feedID, entryHashes); err != nil {
			slog.Error("Unable to cleanup entries",
				slog.Int64("user_id", userID),
				slog.Int64("feed_id", feedID),
				slog.Any("error", err),
			)
		}
	}()

	return newEntries, nil
}

// ArchiveEntries changes the status of entries to "removed" after the given number of days.
func (s *Storage) ArchiveEntries(status string, days, limit int) (int64, error) {
	if days < 0 || limit <= 0 {
		return 0, nil
	}

	query := `
		UPDATE
			entries
		SET
			status=$1
		WHERE
			id IN (
				SELECT
					id
				FROM
					entries
				WHERE
					status=$2 AND
					starred is false AND
					share_code='' AND
					created_at < now () - $3::interval
				ORDER BY
					created_at ASC LIMIT $4
				)
	`

	result, err := s.db.Exec(query, model.EntryStatusRemoved, status, fmt.Sprintf("%d days", days), limit)
	if err != nil {
		return 0, fmt.Errorf(`store: unable to archive %s entries: %v`, status, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(`store: unable to get the number of rows affected: %v`, err)
	}

	return count, nil
}

// SetEntriesStatus update the status of the given list of entries.
func (s *Storage) SetEntriesStatus(userID int64, entryIDs []int64, status string) error {
	query := `UPDATE entries SET status=$1, changed_at=now() WHERE user_id=$2 AND id=ANY($3)`
	if _, err := s.db.Exec(query, status, userID, pq.Array(entryIDs)); err != nil {
		return fmt.Errorf(`store: unable to update entries statuses %v: %v`, entryIDs, err)
	}

	return nil
}

func (s *Storage) SetEntriesStatusCount(userID int64, entryIDs []int64, status string) (int, error) {
	if err := s.SetEntriesStatus(userID, entryIDs, status); err != nil {
		return 0, err
	}

	query := `
		SELECT count(*)
		FROM entries e
		    JOIN feeds f ON (f.id = e.feed_id)
		    JOIN categories c ON (c.id = f.category_id)
		WHERE e.user_id = $1
			AND e.id = ANY($2)
			AND NOT f.hide_globally
			AND NOT c.hide_globally
	`
	row := s.db.QueryRow(query, userID, pq.Array(entryIDs))
	visible := 0
	if err := row.Scan(&visible); err != nil {
		return 0, fmt.Errorf(`store: unable to query entries visibility %v: %v`, entryIDs, err)
	}

	return visible, nil
}

// SetEntriesBookmarked update the bookmarked state for the given list of entries.
func (s *Storage) SetEntriesBookmarkedState(userID int64, entryIDs []int64, starred bool) error {
	query := `UPDATE entries SET starred=$1, changed_at=now() WHERE user_id=$2 AND id=ANY($3)`
	result, err := s.db.Exec(query, starred, userID, pq.Array(entryIDs))
	if err != nil {
		return fmt.Errorf(`store: unable to update the bookmarked state %v: %v`, entryIDs, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(`store: unable to update these entries %v: %v`, entryIDs, err)
	}

	if count == 0 {
		return errors.New(`store: nothing has been updated`)
	}

	return nil
}

// ToggleBookmark toggles entry bookmark value.
func (s *Storage) ToggleBookmark(userID int64, entryID int64) error {
	query := `UPDATE entries SET starred = NOT starred, changed_at=now() WHERE user_id=$1 AND id=$2`
	result, err := s.db.Exec(query, userID, entryID)
	if err != nil {
		return fmt.Errorf(`store: unable to toggle bookmark flag for entry #%d: %v`, entryID, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(`store: unable to toggle bookmark flag for entry #%d: %v`, entryID, err)
	}

	if count == 0 {
		return errors.New(`store: nothing has been updated`)
	}

	return nil
}

// FlushHistory changes all entries with the status "read" to "removed".
func (s *Storage) FlushHistory(userID int64) error {
	query := `
		UPDATE
			entries
		SET
			status=$1,
			changed_at=now()
		WHERE
			user_id=$2 AND status=$3 AND starred is false AND share_code=''
	`
	_, err := s.db.Exec(query, model.EntryStatusRemoved, userID, model.EntryStatusRead)
	if err != nil {
		return fmt.Errorf(`store: unable to flush history: %v`, err)
	}

	return nil
}

// MarkAllAsRead updates all user entries to the read status.
func (s *Storage) MarkAllAsRead(userID int64) error {
	query := `UPDATE entries SET status=$1, changed_at=now() WHERE user_id=$2 AND status=$3`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, model.EntryStatusUnread)
	if err != nil {
		return fmt.Errorf(`store: unable to mark all entries as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	slog.Debug("Marked all entries as read",
		slog.Int64("user_id", userID),
		slog.Int64("nb_entries", count),
	)

	return nil
}

// MarkAllAsReadBeforeDate updates all user entries to the read status before the given date.
func (s *Storage) MarkAllAsReadBeforeDate(userID int64, before time.Time) error {
	query := `
		UPDATE
			entries
		SET
			status=$1,
			changed_at=now()
		WHERE
			user_id=$2 AND status=$3 AND published_at < $4
	`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, model.EntryStatusUnread, before)
	if err != nil {
		return fmt.Errorf(`store: unable to mark all entries as read before %s: %v`, before.Format(time.RFC3339), err)
	}
	count, _ := result.RowsAffected()
	slog.Debug("Marked all entries as read before date",
		slog.Int64("user_id", userID),
		slog.Int64("nb_entries", count),
		slog.String("before", before.Format(time.RFC3339)),
	)
	return nil
}

// MarkGloballyVisibleFeedsAsRead updates all user entries to the read status.
func (s *Storage) MarkGloballyVisibleFeedsAsRead(userID int64) error {
	query := `
		UPDATE
			entries
		SET
			status=$1,
			changed_at=now()
		FROM
			feeds
		WHERE
			entries.feed_id = feeds.id
			AND entries.user_id=$2
			AND entries.status=$3
			AND feeds.hide_globally=$4
	`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, model.EntryStatusUnread, false)
	if err != nil {
		return fmt.Errorf(`store: unable to mark globally visible feeds as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	slog.Debug("Marked globally visible feed entries as read",
		slog.Int64("user_id", userID),
		slog.Int64("nb_entries", count),
	)

	return nil
}

// MarkFeedAsRead updates all feed entries to the read status.
func (s *Storage) MarkFeedAsRead(userID, feedID int64, before time.Time) error {
	query := `
		UPDATE
			entries
		SET
			status=$1,
			changed_at=now()
		WHERE
			user_id=$2 AND feed_id=$3 AND status=$4 AND published_at < $5
	`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, feedID, model.EntryStatusUnread, before)
	if err != nil {
		return fmt.Errorf(`store: unable to mark feed entries as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	slog.Debug("Marked feed entries as read",
		slog.Int64("user_id", userID),
		slog.Int64("feed_id", feedID),
		slog.Int64("nb_entries", count),
		slog.String("before", before.Format(time.RFC3339)),
	)

	return nil
}

// MarkCategoryAsRead updates all category entries to the read status.
func (s *Storage) MarkCategoryAsRead(userID, categoryID int64, before time.Time) error {
	query := `
		UPDATE
			entries
		SET
			status=$1,
			changed_at=now()
		FROM
			feeds
		WHERE
			feed_id=feeds.id
		AND
			feeds.user_id=$2
		AND
			status=$3
		AND
			published_at < $4
		AND
			feeds.category_id=$5
	`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, model.EntryStatusUnread, before, categoryID)
	if err != nil {
		return fmt.Errorf(`store: unable to mark category entries as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	slog.Debug("Marked category entries as read",
		slog.Int64("user_id", userID),
		slog.Int64("category_id", categoryID),
		slog.Int64("nb_entries", count),
		slog.String("before", before.Format(time.RFC3339)),
	)

	return nil
}

// EntryShareCode returns the share code of the provided entry.
// It generates a new one if not already defined.
func (s *Storage) EntryShareCode(userID int64, entryID int64) (shareCode string, err error) {
	query := `SELECT share_code FROM entries WHERE user_id=$1 AND id=$2`
	err = s.db.QueryRow(query, userID, entryID).Scan(&shareCode)
	if err != nil {
		err = fmt.Errorf(`store: unable to get share code for entry #%d: %v`, entryID, err)
		return
	}

	if shareCode == "" {
		shareCode = crypto.GenerateRandomStringHex(20)

		query = `UPDATE entries SET share_code = $1 WHERE user_id=$2 AND id=$3`
		_, err = s.db.Exec(query, shareCode, userID, entryID)
		if err != nil {
			err = fmt.Errorf(`store: unable to set share code for entry #%d: %v`, entryID, err)
			return
		}
	}

	return
}

// UnshareEntry removes the share code for the given entry.
func (s *Storage) UnshareEntry(userID int64, entryID int64) (err error) {
	query := `UPDATE entries SET share_code='' WHERE user_id=$1 AND id=$2`
	_, err = s.db.Exec(query, userID, entryID)
	if err != nil {
		err = fmt.Errorf(`store: unable to remove share code for entry #%d: %v`, entryID, err)
	}
	return
}

// truncateStringForTSVectorField truncates a string to fit within the maximum size for a TSVector field in PostgreSQL.
func truncateStringForTSVectorField(s string) string {
	// The length of a tsvector (lexemes + positions) must be less than 1 megabyte.
	const maxTSVectorSize = 1024 * 1024

	if len(s) < maxTSVectorSize {
		return s
	}

	// Truncate to fit under the limit, ensuring we don't break UTF-8 characters
	truncated := s[:maxTSVectorSize-1]

	// Walk backwards to find the last complete UTF-8 character
	for i := len(truncated) - 1; i >= 0; i-- {
		if (truncated[i] & 0x80) == 0 {
			// ASCII character, we can stop here
			return truncated[:i+1]
		}
		if (truncated[i] & 0xC0) == 0xC0 {
			// Start of a multi-byte UTF-8 character
			return truncated[:i]
		}
	}

	// Fallback: return empty string if we can't find a valid UTF-8 boundary
	return ""
}
