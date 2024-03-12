// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"strings"

	"miniflux.app/v2/internal/model"

	"github.com/lib/pq"
)

// GetEnclosures returns all attachments for the given entry.
func (s *Storage) GetEnclosures(entryID int64) (model.EnclosureList, error) {
	query := `
		SELECT
			id,
			user_id,
			entry_id,
			url,
			size,
			mime_type,
		    media_progression
		FROM
			enclosures
		WHERE
			entry_id = $1
		ORDER BY id ASC
	`

	rows, err := s.db.Query(query, entryID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch enclosures: %v`, err)
	}
	defer rows.Close()

	enclosures := make(model.EnclosureList, 0)
	for rows.Next() {
		var enclosure model.Enclosure
		err := rows.Scan(
			&enclosure.ID,
			&enclosure.UserID,
			&enclosure.EntryID,
			&enclosure.URL,
			&enclosure.Size,
			&enclosure.MimeType,
			&enclosure.MediaProgression,
		)

		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch enclosure row: %v`, err)
		}

		enclosures = append(enclosures, &enclosure)
	}

	return enclosures, nil
}

func (s *Storage) GetEnclosure(enclosureID int64) (*model.Enclosure, error) {
	query := `
		SELECT
			id,
			user_id,
			entry_id,
			url,
			size,
			mime_type,
		    media_progression
		FROM
			enclosures
		WHERE
			id = $1
		ORDER BY id ASC
	`

	row := s.db.QueryRow(query, enclosureID)

	var enclosure model.Enclosure
	err := row.Scan(
		&enclosure.ID,
		&enclosure.UserID,
		&enclosure.EntryID,
		&enclosure.URL,
		&enclosure.Size,
		&enclosure.MimeType,
		&enclosure.MediaProgression,
	)

	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch enclosure row: %v`, err)
	}

	return &enclosure, nil
}

func (s *Storage) createEnclosure(tx *sql.Tx, enclosure *model.Enclosure) error {
	enclosureURL := strings.TrimSpace(enclosure.URL)
	if enclosureURL == "" {
		return nil
	}

	query := `
		INSERT INTO enclosures
			(url, size, mime_type, entry_id, user_id, media_progression)
		VALUES
			($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, entry_id, md5(url)) DO NOTHING
		RETURNING
			id
	`
	if err := tx.QueryRow(
		query,
		enclosureURL,
		enclosure.Size,
		enclosure.MimeType,
		enclosure.EntryID,
		enclosure.UserID,
		enclosure.MediaProgression,
	).Scan(&enclosure.ID); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf(`store: unable to create enclosure: %w`, err)
	}

	return nil
}

func (s *Storage) updateEnclosures(tx *sql.Tx, entry *model.Entry) error {
	if len(entry.Enclosures) == 0 {
		return nil
	}

	sqlValues := make([]string, 0, len(entry.Enclosures))
	for _, enclosure := range entry.Enclosures {
		sqlValues = append(sqlValues, strings.TrimSpace(enclosure.URL))

		if err := s.createEnclosure(tx, enclosure); err != nil {
			return err
		}
	}

	query := `
		DELETE FROM
			enclosures
		WHERE
			user_id=$1 AND entry_id=$2 AND url <> ALL($3)
	`

	_, err := tx.Exec(query, entry.UserID, entry.ID, pq.Array(sqlValues))
	if err != nil {
		return fmt.Errorf(`store: unable to delete old enclosures: %v`, err)
	}

	return nil
}

func (s *Storage) UpdateEnclosure(enclosure *model.Enclosure) error {
	query := `
		UPDATE
			enclosures
		SET
			url=$1,
			size=$2,
			mime_type=$3,
			entry_id=$4, 
			user_id=$5, 
			media_progression=$6
		WHERE
			id=$7
	`
	_, err := s.db.Exec(query,
		enclosure.URL,
		enclosure.Size,
		enclosure.MimeType,
		enclosure.EntryID,
		enclosure.UserID,
		enclosure.MediaProgression,
		enclosure.ID,
	)

	if err != nil {
		return fmt.Errorf(`store: unable to update enclosure #%d : %v`, enclosure.ID, err)
	}

	return nil
}
