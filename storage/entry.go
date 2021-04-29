// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"miniflux.app/crypto"
	"miniflux.app/logger"
	"miniflux.app/model"

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

	n, err := builder.CountEntries()
	if err != nil {
		logger.Error(`store: unable to count unread entries for user #%d: %v`, userID, err)
		return 0
	}

	return n
}

// NewEntryQueryBuilder returns a new EntryQueryBuilder
func (s *Storage) NewEntryQueryBuilder(userID int64) *EntryQueryBuilder {
	return NewEntryQueryBuilder(s, userID)
}

// UpdateEntryContent updates entry content.
func (s *Storage) UpdateEntryContent(entry *model.Entry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	query := `
		UPDATE
			entries
		SET
			content=$1, reading_time=$2
		WHERE
			id=$3 AND user_id=$4
	`
	_, err = tx.Exec(query, entry.Content, entry.ReadingTime, entry.ID, entry.UserID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(`store: unable to update content of entry #%d: %v`, entry.ID, err)
	}

	query = `
		UPDATE
			entries
		SET
			document_vectors = setweight(to_tsvector(left(coalesce(title, ''), 500000)), 'A') || setweight(to_tsvector(left(coalesce(content, ''), 500000)), 'B')
		WHERE
			id=$1 AND user_id=$2
	`
	_, err = tx.Exec(query, entry.ID, entry.UserID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(`store: unable to update content of entry #%d: %v`, entry.ID, err)
	}

	return tx.Commit()
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
				document_vectors
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
				setweight(to_tsvector(left(coalesce($1, ''), 500000)), 'A') || setweight(to_tsvector(left(coalesce($6, ''), 500000)), 'B')
			)
		RETURNING
			id, status
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
	).Scan(&entry.ID, &entry.Status)

	if err != nil {
		return fmt.Errorf(`store: unable to create entry %q (feed #%d): %v`, entry.URL, entry.FeedID, err)
	}

	for i := 0; i < len(entry.Enclosures); i++ {
		entry.Enclosures[i].EntryID = entry.ID
		entry.Enclosures[i].UserID = entry.UserID
		err := s.createEnclosure(tx, entry.Enclosures[i])
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
			document_vectors = setweight(to_tsvector(left(coalesce($1, ''), 500000)), 'A') || setweight(to_tsvector(left(coalesce($4, ''), 500000)), 'B')
		WHERE
			user_id=$7 AND feed_id=$8 AND hash=$9
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
		entry.UserID,
		entry.FeedID,
		entry.Hash,
	).Scan(&entry.ID)

	if err != nil {
		return fmt.Errorf(`store: unable to update entry %q: %v`, entry.URL, err)
	}

	for _, enclosure := range entry.Enclosures {
		enclosure.UserID = entry.UserID
		enclosure.EntryID = entry.ID
	}

	return s.updateEnclosures(tx, entry.UserID, entry.ID, entry.Enclosures)
}

// entryExists checks if an entry already exists based on its hash when refreshing a feed.
func (s *Storage) entryExists(tx *sql.Tx, entry *model.Entry) bool {
	var result bool
	tx.QueryRow(
		`SELECT true FROM entries WHERE user_id=$1 AND feed_id=$2 AND hash=$3`,
		entry.UserID,
		entry.FeedID,
		entry.Hash,
	).Scan(&result)
	return result
}

// GetReadTime fetches the read time of an entry based on its hash, and the feed id and user id from the feed.
// It's intended to be used on entries objects created by parsing a feed as they don't contain much information.
// The feed param helps to scope the search to a specific user and feed in order to avoid hash clashes.
func (s *Storage) GetReadTime(entry *model.Entry, feed *model.Feed) int {
	var result int
	s.db.QueryRow(
		`SELECT reading_time FROM entries WHERE user_id=$1 AND feed_id=$2 AND hash=$3`,
		feed.UserID,
		feed.ID,
		entry.Hash,
	).Scan(&result)
	return result
}

// cleanupEntries deletes from the database entries marked as "removed" and not visible anymore in the feed.
func (s *Storage) cleanupEntries(feedID int64, entryHashes []string) error {
	query := `
		DELETE FROM
			entries
		WHERE
			feed_id=$1
		AND
			id IN (SELECT id FROM entries WHERE feed_id=$2 AND status=$3 AND NOT (hash=ANY($4)))
	`
	if _, err := s.db.Exec(query, feedID, feedID, model.EntryStatusRemoved, pq.Array(entryHashes)); err != nil {
		return fmt.Errorf(`store: unable to cleanup entries: %v`, err)
	}

	return nil
}

// RefreshFeedEntries updates feed entries while refreshing a feed.
func (s *Storage) RefreshFeedEntries(userID, feedID int64, entries model.Entries, updateExistingEntries bool) (err error) {
	var entryHashes []string

	for _, entry := range entries {
		entry.UserID = userID
		entry.FeedID = feedID

		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf(`store: unable to start transaction: %v`, err)
		}

		if s.entryExists(tx, entry) {
			if updateExistingEntries {
				err = s.updateEntry(tx, entry)
			}
		} else {
			err = s.createEntry(tx, entry)
		}

		if err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf(`store: unable to commit transaction: %v`, err)
		}

		entryHashes = append(entryHashes, entry.Hash)
	}

	go func() {
		if err := s.cleanupEntries(feedID, entryHashes); err != nil {
			logger.Error(`store: feed #%d: %v`, feedID, err)
		}
	}()

	return nil
}

// ArchiveEntries changes the status of entries to "removed" after the given number of days.
func (s *Storage) ArchiveEntries(status string, days int) (int64, error) {
	if days < 0 {
		return 0, nil
	}

	query := `
		UPDATE
			entries
		SET
			status='removed'
		WHERE
			id=ANY(SELECT id FROM entries WHERE status=$1 AND starred is false AND share_code='' AND created_at < now () - '%d days'::interval ORDER BY created_at ASC LIMIT 5000)
	`

	result, err := s.db.Exec(fmt.Sprintf(query, days), status)
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
	if status == model.EntryStatusRead {
		return s.markEntriesAsRead(userID, entryIDs)
	}

	query := `UPDATE entries SET status=$1, changed_at=now() WHERE user_id=$2 AND id=ANY($3)`
	result, err := s.db.Exec(query, status, userID, pq.Array(entryIDs))
	if err != nil {
		return fmt.Errorf(`store: unable to update entries statuses %v: %v`, entryIDs, err)
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

// markEntriesAsRead updates entries to the read status,
// this function also updates read_at which SetEntriesStatus would not do.
func (s *Storage) markEntriesAsRead(userID int64, entryIDs []int64) error {
	query := `
		UPDATE
			entries
		SET
			status=$1,
			changed_at=now(),
			read_at=now()
		WHERE
			user_id=$2 AND id=ANY($3)
	`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, pq.Array(entryIDs))
	if err != nil {
		return fmt.Errorf(`store: unable to mark feed entries as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	logger.Debug("[Storage:MarkFeedAsRead] %d items marked as read", count)

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

// FlushHistory set all entries with the status "read" to "removed".
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
	query := `UPDATE entries SET status=$1, changed_at=now(), read_at=now() WHERE user_id=$2 AND status=$3`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, model.EntryStatusUnread)
	if err != nil {
		return fmt.Errorf(`store: unable to mark all entries as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	logger.Debug("[Storage:MarkAllAsRead] %d items marked as read", count)

	return nil
}

// MarkFeedAsRead updates all feed entries to the read status.
func (s *Storage) MarkFeedAsRead(userID, feedID int64, before time.Time) error {
	query := `
		UPDATE
			entries
		SET
			status=$1,
			changed_at=now(),
			read_at=now()
		WHERE
			user_id=$2 AND feed_id=$3 AND status=$4 AND published_at < $5
	`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, feedID, model.EntryStatusUnread, before)
	if err != nil {
		return fmt.Errorf(`store: unable to mark feed entries as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	logger.Debug("[Storage:MarkFeedAsRead] %d items marked as read", count)

	return nil
}

// MarkCategoryAsRead updates all category entries to the read status.
func (s *Storage) MarkCategoryAsRead(userID, categoryID int64, before time.Time) error {
	query := `
		UPDATE
			entries
		SET
			status=$1,
			changed_at=now(),
			read_at=now()
		WHERE
			user_id=$2
		AND
			status=$3
		AND
			published_at < $4
		AND
			feed_id IN (SELECT id FROM feeds WHERE user_id=$2 AND category_id=$5)
	`
	result, err := s.db.Exec(query, model.EntryStatusRead, userID, model.EntryStatusUnread, before, categoryID)
	if err != nil {
		return fmt.Errorf(`store: unable to mark category entries as read: %v`, err)
	}

	count, _ := result.RowsAffected()
	logger.Debug("[Storage:MarkCategoryAsRead] %d items marked as read", count)

	return nil
}

// EntryURLExists returns true if an entry with this URL already exists.
func (s *Storage) EntryURLExists(feedID int64, entryURL string) bool {
	var result bool
	query := `SELECT true FROM entries WHERE feed_id=$1 AND url=$2`
	s.db.QueryRow(query, feedID, entryURL).Scan(&result)
	return result
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
