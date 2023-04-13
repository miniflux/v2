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
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf(`store: unable to start transaction: %v`, err)
	}
	// As the transaction is only created to use the txGetEnclosures function, we can commit it and close it.
	// to avoid leaving an open transaction as I don't have any idea if it will be closed automatically,
	// I manually close it. I chose `commit` over `rollback` because I assumed it cost less on SGBD, but I'm no Database
	// administrator so any better solution is welcome.
	defer tx.Commit()
	return s.txGetEnclosures(tx, entryID)
}

// GetEnclosures returns all attachments for the given entry within a Database transaction
func (s *Storage) txGetEnclosures(tx *sql.Tx, entryID int64) (model.EnclosureList, error) {
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

	rows, err := tx.Query(query, entryID)
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
	if enclosure.URL == "" {
		return nil
	}

	query := `
		INSERT INTO enclosures
			(url, size, mime_type, entry_id, user_id, media_progression)
		VALUES
			($1, $2, $3, $4, $5, $6)
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
		enclosure.MediaProgression,
	).Scan(&enclosure.ID)

	if err != nil {
		return fmt.Errorf(`store: unable to create enclosure %q: %v`, enclosure.URL, err)
	}

	return nil
}

func (s *Storage) updateEnclosures(tx *sql.Tx, userID, entryID int64, enclosures model.EnclosureList) error {
	originalEnclosures, err := s.txGetEnclosures(tx, entryID)
	if err != nil {
		return fmt.Errorf(`store: unable fetch enclosures for entry #%d : %v`, entryID, err)
	}

	// this map will allow to identify enclosure already in the database based on their URL.
	originalEnclosuresByURL := map[string]*model.Enclosure{}
	for _, enclosure := range originalEnclosures {
		originalEnclosuresByURL[enclosure.URL] = enclosure
	}

	// in order to keep enclosure ID consistent I need to identify already existing one to keep them as is, and only
	// add/delete enclosure that need to be
	enclosuresToAdd := map[string]*model.Enclosure{}
	enclosuresToDelete := map[string]*model.Enclosure{}
	enclosuresToKeep := map[string]*model.Enclosure{}

	for _, enclosure := range enclosures {
		originalEnclosure, alreadyExist := originalEnclosuresByURL[enclosure.URL]
		if alreadyExist {
			enclosuresToKeep[originalEnclosure.URL] = originalEnclosure // we keep the original already in the database
		} else {
			enclosuresToAdd[enclosure.URL] = enclosure // we insert the new one
		}
	}

	// we know what to keep, and add. We need to find what's in the database that need to be deleted
	for _, enclosure := range originalEnclosures {
		_, existToAdd := enclosuresToAdd[enclosure.URL]
		_, existToKeep := enclosuresToKeep[enclosure.URL]
		if !existToKeep && !existToAdd { // if it does not exist to keep or add this mean it has been deleted.
			enclosuresToDelete[enclosure.URL] = enclosure
		}
	}

	for _, enclosure := range enclosuresToDelete {
		if _, err := tx.Exec(`DELETE FROM enclosures WHERE user_id=$1 AND entry_id=$2 and id=$3`, userID, entryID, enclosure.ID); err != nil {
			return err
		}
	}

	for _, enclosure := range enclosuresToAdd {
		if err := s.createEnclosure(tx, enclosure); err != nil {
			return err
		}
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
