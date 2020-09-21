// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"fmt"

	"miniflux.app/model"
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
			mime_type
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
		)

		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch enclosure row: %v`, err)
		}

		enclosures = append(enclosures, &enclosure)
	}

	return enclosures, nil
}

func (s *Storage) createEnclosure(tx *sql.Tx, enclosure *model.Enclosure) error {
	if enclosure.URL == "" {
		return nil
	}

	query := `
		INSERT INTO enclosures
			(url, size, mime_type, entry_id, user_id)
		VALUES
			($1, $2, $3, $4, $5)
		RETURNING
			id
	`
	err := tx.QueryRow(
		query,
		enclosure.URL,
		enclosure.Size,
		enclosure.MimeType,
		enclosure.EntryID,
		enclosure.UserID,
	).Scan(&enclosure.ID)

	if err != nil {
		return fmt.Errorf(`store: unable to create enclosure %q: %v`, enclosure.URL, err)
	}

	return nil
}

func (s *Storage) updateEnclosures(tx *sql.Tx, userID, entryID int64, enclosures model.EnclosureList) error {
	// We delete all attachments in the transaction to keep only the ones visible in the feeds.
	if _, err := tx.Exec(`DELETE FROM enclosures WHERE user_id=$1 AND entry_id=$2`, userID, entryID); err != nil {
		return err
	}

	for _, enclosure := range enclosures {
		if err := s.createEnclosure(tx, enclosure); err != nil {
			return err
		}
	}

	return nil
}
