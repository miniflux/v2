// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"net/http"
	"strconv"
	"strings"
	"time"

	"miniflux.app/crypto"
	"miniflux.app/logger"
	"miniflux.app/model"

	"github.com/lib/pq"
)

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
			content=$1
		WHERE
			id=$2 AND user_id=$3
	`
	_, err = tx.Exec(query, entry.Content, entry.ID, entry.UserID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(`store: unable to update content of entry #%d: %v`, entry.ID, err)
	}

	query = `
		UPDATE
			entries
		SET
			document_vectors = setweight(to_tsvector(substring(coalesce(title, '') for 1000000)), 'A') || setweight(to_tsvector(substring(coalesce(content, '') for 1000000)), 'B')
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
func (s *Storage) createEntry(entry *model.Entry) error {
	query := `
		INSERT INTO entries
			(title, hash, url, comments_url, published_at, content, author, user_id, feed_id, changed_at, document_vectors)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), setweight(to_tsvector(substring(coalesce($1, '') for 1000000)), 'A') || setweight(to_tsvector(substring(coalesce($6, '') for 1000000)), 'B'))
		RETURNING
			id, status
	`
	err := s.db.QueryRow(
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
	).Scan(&entry.ID, &entry.Status)

	if err != nil {
		return fmt.Errorf(`store: unable to create entry %q (feed #%d): %v`, entry.URL, entry.FeedID, err)
	}

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

// updateEntry updates an entry when a feed is refreshed.
// Note: we do not update the published date because some feeds do not contains any date,
// it default to time.Now() which could change the order of items on the history page.
func (s *Storage) updateEntry(entry *model.Entry) error {
	query := `
		UPDATE
			entries
		SET
			title=$1,
			url=$2,
			comments_url=$3,
			content=$4,
			author=$5,
			document_vectors = setweight(to_tsvector(substring(coalesce($1, '') for 1000000)), 'A') || setweight(to_tsvector(substring(coalesce($4, '') for 1000000)), 'B')
		WHERE
			user_id=$6 AND feed_id=$7 AND hash=$8
		RETURNING
			id
	`
	err := s.db.QueryRow(
		query,
		entry.Title,
		entry.URL,
		entry.CommentsURL,
		entry.Content,
		entry.Author,
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

	return s.UpdateEnclosures(entry.Enclosures)
}

// entryExists checks if an entry already exists based on its hash when refreshing a feed.
func (s *Storage) entryExists(entry *model.Entry) bool {
	var result int
	query := `SELECT 1 FROM entries WHERE user_id=$1 AND feed_id=$2 AND hash=$3`
	s.db.QueryRow(query, entry.UserID, entry.FeedID, entry.Hash).Scan(&result)
	return result == 1
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

// UpdateEntries updates a list of entries while refreshing a feed.
func (s *Storage) UpdateEntries(userID, feedID int64, entries model.Entries, updateExistingEntries bool) (err error) {
	var entryHashes []string
	var telegramItemMsg []string
	for _, entry := range entries {
		entry.UserID = userID
		entry.FeedID = feedID

		if s.entryExists(entry) {
			if updateExistingEntries {
				err = s.updateEntry(entry)
			}
		} else {
			err = s.createEntry(entry)

			tempText := fmt.Sprintf("[%v](%v)", entry.Title, entry.URL)
			telegramItemMsg = append(telegramItemMsg, tempText)
		}

		if err != nil {
			return err
		}

		entryHashes = append(entryHashes, entry.Hash)
	}

	sendTelegramMsg(userID, feedID, telegramItemMsg, s)

	if err := s.cleanupEntries(feedID, entryHashes); err != nil {
		logger.Error(`store: feed #%d: %v`, feedID, err)
	}

	return nil
}

func sendTelegramMsg(userID int64, feedID int64, telegramItemMsg []string, s *Storage) {
	if len(telegramItemMsg) > 0 {
		integration, _ := s.Integration(userID)
		if integration != nil && integration.TelegramEnabled && len(integration.TelegramToken) > 0 {
			feed, _ := s.FeedByID(userID, feedID)
			bot, _ := tgbotapi.NewBotAPIWithClient(integration.TelegramToken, &http.Client{Timeout: 15 * time.Second})
			if bot != nil {
				text := fmt.Sprintf("*%v*\n\n", feed.Title) + strings.Join(telegramItemMsg, "\n")
				chatID, _ := strconv.ParseInt(integration.TelegramChatID, 10, 64)
				message := tgbotapi.NewMessage(chatID, text)
				message.DisableWebPagePreview = true
				message.ParseMode = "markdown"
				_, err := bot.Send(message)
				if err != nil {
					logger.Error(`telegram: send msg error %v`, feedID, err)
				}
			}
		}
	}
}

// ArchiveEntries changes the status of read items to "removed" after specified days.
func (s *Storage) ArchiveEntries(days int) error {
	if days < 0 {
		return nil
	}

	before := time.Now().AddDate(0, 0, -days)
	query := `
		UPDATE
			entries
		SET
			status=$1
		WHERE
			id=ANY(SELECT id FROM entries WHERE status=$2 AND starred is false AND share_code='' AND published_at < $3 LIMIT 5000)
	`
	if _, err := s.db.Exec(query, model.EntryStatusRemoved, model.EntryStatusRead, before); err != nil {
		return fmt.Errorf(`store: unable to archive read entries: %v`, err)
	}

	return nil
}

// SetEntriesStatus update the status of the given list of entries.
func (s *Storage) SetEntriesStatus(userID int64, entryIDs []int64, status string) error {
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
	query := `UPDATE entries SET status=$1, changed_at=now() WHERE user_id=$2 AND status=$3`
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
			changed_at=now()
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
			changed_at=now()
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
