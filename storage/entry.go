// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"

	"github.com/lib/pq"
)

// GetEntryQueryBuilder returns a new EntryQueryBuilder
func (s *Storage) GetEntryQueryBuilder(userID int64, timezone string) *EntryQueryBuilder {
	return NewEntryQueryBuilder(s, userID, timezone)
}

// CreateEntry add a new entry.
func (s *Storage) CreateEntry(entry *model.Entry) error {
	query := `
		INSERT INTO entries
		(title, hash, url, published_at, content, author, user_id, feed_id)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		entry.Title,
		entry.Hash,
		entry.URL,
		entry.Date,
		entry.Content,
		entry.Author,
		entry.UserID,
		entry.FeedID,
	).Scan(&entry.ID)

	if err != nil {
		return fmt.Errorf("unable to create entry: %v", err)
	}

	entry.Status = "unread"
	for i := 0; i < len(entry.Enclosures); i++ {
		entry.Enclosures[i].EntryID = entry.ID
		entry.Enclosures[i].UserID = entry.UserID
		err := s.CreateEnclosure(entry.Enclosures[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateEntryContent updates entry content.
func (s *Storage) UpdateEntryContent(entry *model.Entry) error {
	query := `
		UPDATE entries SET
		content=$1
		WHERE user_id=$2 AND id=$3
	`

	_, err := s.db.Exec(
		query,
		entry.Content,
		entry.UserID,
		entry.ID,
	)
	return err
}

// UpdateEntry update an entry when a feed is refreshed.
func (s *Storage) UpdateEntry(entry *model.Entry) error {
	query := `
		UPDATE entries SET
		title=$1, url=$2, published_at=$3, content=$4, author=$5
		WHERE user_id=$6 AND feed_id=$7 AND hash=$8
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		entry.Title,
		entry.URL,
		entry.Date,
		entry.Content,
		entry.Author,
		entry.UserID,
		entry.FeedID,
		entry.Hash,
	).Scan(&entry.ID)

	if err != nil {
		return err
	}

	for _, enclosure := range entry.Enclosures {
		enclosure.UserID = entry.UserID
		enclosure.EntryID = entry.ID
	}

	return s.UpdateEnclosures(entry.Enclosures)
}

// EntryExists checks if an entry already exists based on its hash when refreshing a feed.
func (s *Storage) EntryExists(entry *model.Entry) bool {
	var result int
	query := `SELECT count(*) as c FROM entries WHERE user_id=$1 AND feed_id=$2 AND hash=$3`
	s.db.QueryRow(query, entry.UserID, entry.FeedID, entry.Hash).Scan(&result)
	return result >= 1
}

// UpdateEntries update a list of entries while refreshing a feed.
func (s *Storage) UpdateEntries(userID, feedID int64, entries model.Entries) (err error) {
	var entryHashes []string
	for _, entry := range entries {
		entry.UserID = userID
		entry.FeedID = feedID

		if s.EntryExists(entry) {
			err = s.UpdateEntry(entry)
		} else {
			err = s.CreateEntry(entry)
		}

		if err != nil {
			return err
		}

		entryHashes = append(entryHashes, entry.Hash)
	}

	if err := s.CleanupEntries(feedID, entryHashes); err != nil {
		log.Println(err)
	}

	return nil
}

// CleanupEntries deletes from the database entries marked as "removed" and not visible anymore in the feed.
func (s *Storage) CleanupEntries(feedID int64, entryHashes []string) error {
	query := `
		DELETE FROM entries
		WHERE feed_id=$1 AND
		id IN (SELECT id FROM entries WHERE feed_id=$2 AND status=$3 AND NOT (hash=ANY($4)))
	`
	if _, err := s.db.Exec(query, feedID, feedID, model.EntryStatusRemoved, pq.Array(entryHashes)); err != nil {
		return fmt.Errorf("unable to cleanup entries: %v", err)
	}

	return nil
}

// SetEntriesStatus update the status of the given list of entries.
func (s *Storage) SetEntriesStatus(userID int64, entryIDs []int64, status string) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:SetEntriesStatus] userID=%d, entryIDs=%v, status=%s", userID, entryIDs, status))

	query := `UPDATE entries SET status=$1 WHERE user_id=$2 AND id=ANY($3)`
	result, err := s.db.Exec(query, status, userID, pq.Array(entryIDs))
	if err != nil {
		return fmt.Errorf("unable to update entries status: %v", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to update these entries: %v", err)
	}

	if count == 0 {
		return errors.New("nothing has been updated")
	}

	return nil
}

// FlushHistory set all entries with the status "read" to "removed".
func (s *Storage) FlushHistory(userID int64) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:FlushHistory] userID=%d", userID))

	query := `UPDATE entries SET status=$1 WHERE user_id=$2 AND status=$3`
	_, err := s.db.Exec(query, model.EntryStatusRemoved, userID, model.EntryStatusRead)
	if err != nil {
		return fmt.Errorf("unable to flush history: %v", err)
	}

	return nil
}
