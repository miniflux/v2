// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"errors"
	"fmt"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"time"

	"github.com/lib/pq"
)

func (s *Storage) GetEntryQueryBuilder(userID int64, timezone string) *EntryQueryBuilder {
	return NewEntryQueryBuilder(s, userID, timezone)
}

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
		return fmt.Errorf("Unable to create entry: %v", err)
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

func (s *Storage) UpdateEntry(entry *model.Entry) error {
	query := `
		UPDATE entries SET
		title=$1, url=$2, published_at=$3, content=$4, author=$5
		WHERE user_id=$6 AND feed_id=$7 AND hash=$8
	`
	_, err := s.db.Exec(
		query,
		entry.Title,
		entry.URL,
		entry.Date,
		entry.Content,
		entry.Author,
		entry.UserID,
		entry.FeedID,
		entry.Hash,
	)

	return err
}

func (s *Storage) EntryExists(entry *model.Entry) bool {
	var result int
	query := `SELECT count(*) as c FROM entries WHERE user_id=$1 AND feed_id=$2 AND hash=$3`
	s.db.QueryRow(query, entry.UserID, entry.FeedID, entry.Hash).Scan(&result)
	return result >= 1
}

func (s *Storage) UpdateEntries(userID, feedID int64, entries model.Entries) (err error) {
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
	}

	return nil
}

func (s *Storage) SetEntriesStatus(userID int64, entryIDs []int64, status string) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:SetEntriesStatus] userID=%d, entryIDs=%v, status=%s", userID, entryIDs, status))

	query := `UPDATE entries SET status=$1 WHERE user_id=$2 AND id=ANY($3)`
	result, err := s.db.Exec(query, status, userID, pq.Array(entryIDs))
	if err != nil {
		return fmt.Errorf("Unable to update entry status: %v", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Unable to update this entry: %v", err)
	}

	if count == 0 {
		return errors.New("Nothing has been updated")
	}

	return nil
}
