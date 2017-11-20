// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"github.com/miniflux/miniflux2/model"
)

func (s *Storage) GetEnclosures(entryID int64) (model.EnclosureList, error) {
	query := `SELECT
		id, user_id, entry_id, url, size, mime_type
		FROM enclosures
		WHERE entry_id = $1 ORDER BY id ASC`

	rows, err := s.db.Query(query, entryID)
	if err != nil {
		return nil, fmt.Errorf("Unable to get enclosures: %v", err)
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
			return nil, fmt.Errorf("Unable to fetch enclosure row: %v", err)
		}

		enclosures = append(enclosures, &enclosure)
	}

	return enclosures, nil
}

func (s *Storage) CreateEnclosure(enclosure *model.Enclosure) error {
	query := `
		INSERT INTO enclosures
		(url, size, mime_type, entry_id, user_id)
		VALUES
		($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		enclosure.URL,
		enclosure.Size,
		enclosure.MimeType,
		enclosure.EntryID,
		enclosure.UserID,
	).Scan(&enclosure.ID)

	if err != nil {
		return fmt.Errorf("Unable to create enclosure: %v", err)
	}

	return nil
}
