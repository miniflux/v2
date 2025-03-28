// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"strings"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
)

// HasFeedIcon checks if the given feed has an icon.
func (s *Storage) HasFeedIcon(feedID int64) bool {
	var result bool
	query := `SELECT true FROM feed_icons WHERE feed_id=$1`
	s.db.QueryRow(query, feedID).Scan(&result)
	return result
}

// IconByID returns an icon by the ID.
func (s *Storage) IconByID(iconID int64) (*model.Icon, error) {
	var icon model.Icon
	query := `
		SELECT
			id,
			hash,
			mime_type,
			content,
			external_id
		FROM icons
		WHERE id=$1`
	err := s.db.QueryRow(query, iconID).Scan(&icon.ID, &icon.Hash, &icon.MimeType, &icon.Content, &icon.ExternalID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("store: unable to fetch icon #%d: %w", iconID, err)
	}

	return &icon, nil
}

// IconByExternalID returns an icon by the External Icon ID.
func (s *Storage) IconByExternalID(externalIconID string) (*model.Icon, error) {
	var icon model.Icon
	query := `
		SELECT
			id,
			hash,
			mime_type,
			content,
			external_id
		FROM icons
		WHERE external_id=$1
	`
	err := s.db.QueryRow(query, externalIconID).Scan(&icon.ID, &icon.Hash, &icon.MimeType, &icon.Content, &icon.ExternalID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("store: unable to fetch icon %s: %w", externalIconID, err)
	}

	return &icon, nil
}

// IconByFeedID returns a feed icon.
func (s *Storage) IconByFeedID(userID, feedID int64) (*model.Icon, error) {
	query := `
		SELECT
			icons.id,
			icons.hash,
			icons.mime_type,
			icons.content,
			icons.external_id
		FROM icons
		LEFT JOIN feed_icons ON feed_icons.icon_id=icons.id
		LEFT JOIN feeds ON feeds.id=feed_icons.feed_id
		WHERE
			feeds.user_id=$1 AND feeds.id=$2
		LIMIT 1
	`
	var icon model.Icon
	err := s.db.QueryRow(query, userID, feedID).Scan(&icon.ID, &icon.Hash, &icon.MimeType, &icon.Content, &icon.ExternalID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch icon: %v`, err)
	}

	return &icon, nil
}

// StoreFeedIcon creates or updates a feed icon.
func (s *Storage) StoreFeedIcon(feedID int64, icon *model.Icon) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf(`store: unable to start transaction: %v`, err)
	}

	if err := tx.QueryRow(`SELECT id FROM icons WHERE hash=$1`, icon.Hash).Scan(&icon.ID); err == sql.ErrNoRows {
		query := `
			INSERT INTO icons
				(hash, mime_type, content, external_id)
			VALUES
				($1, $2, $3, $4)
			RETURNING
				id
		`
		err := tx.QueryRow(
			query,
			icon.Hash,
			normalizeMimeType(icon.MimeType),
			icon.Content,
			crypto.GenerateRandomStringHex(20),
		).Scan(&icon.ID)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf(`store: unable to create icon: %v`, err)
		}
	} else if err != nil {
		tx.Rollback()
		return fmt.Errorf(`store: unable to fetch icon by hash %q: %v`, icon.Hash, err)
	}

	if _, err := tx.Exec(`DELETE FROM feed_icons WHERE feed_id=$1`, feedID); err != nil {
		tx.Rollback()
		return fmt.Errorf(`store: unable to delete feed icon: %v`, err)
	}

	if _, err := tx.Exec(`INSERT INTO feed_icons (feed_id, icon_id) VALUES ($1, $2)`, feedID, icon.ID); err != nil {
		tx.Rollback()
		return fmt.Errorf(`store: unable to associate feed and icon: %v`, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(`store: unable to commit transaction: %v`, err)
	}

	return nil
}

// Icons returns all icons that belongs to a user.
func (s *Storage) Icons(userID int64) (model.Icons, error) {
	query := `
		SELECT
			icons.id,
			icons.hash,
			icons.mime_type,
			icons.content,
			icons.external_id
		FROM icons
		LEFT JOIN feed_icons ON feed_icons.icon_id=icons.id
		LEFT JOIN feeds ON feeds.id=feed_icons.feed_id
		WHERE
			feeds.user_id=$1
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch icons: %v`, err)
	}
	defer rows.Close()

	var icons model.Icons
	for rows.Next() {
		var icon model.Icon
		err := rows.Scan(&icon.ID, &icon.Hash, &icon.MimeType, &icon.Content, &icon.ExternalID)
		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch icons row: %v`, err)
		}
		icons = append(icons, &icon)
	}

	return icons, nil
}

func normalizeMimeType(mimeType string) string {
	mimeType = strings.ToLower(mimeType)
	switch mimeType {
	case "image/png", "image/jpeg", "image/jpg", "image/webp", "image/svg+xml", "image/x-icon", "image/gif":
		return mimeType
	default:
		return "image/x-icon"
	}
}
