// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/miniflux/miniflux/helper"
	"github.com/miniflux/miniflux/model"
)

// HasIcon checks if the given feed has an icon.
func (s *Storage) HasIcon(feedID int64) bool {
	var result int
	query := `SELECT count(*) as c FROM feed_icons WHERE feed_id=$1`
	s.db.QueryRow(query, feedID).Scan(&result)
	return result == 1
}

// IconByID returns an icon by the ID.
func (s *Storage) IconByID(iconID int64) (*model.Icon, error) {
	defer helper.ExecutionTime(time.Now(), "[Storage:IconByID]")

	var icon model.Icon
	query := `SELECT id, hash, mime_type, content FROM icons WHERE id=$1`
	err := s.db.QueryRow(query, iconID).Scan(&icon.ID, &icon.Hash, &icon.MimeType, &icon.Content)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("Unable to fetch icon by hash: %v", err)
	}

	return &icon, nil
}

// IconByHash returns an icon by the hash (checksum).
func (s *Storage) IconByHash(icon *model.Icon) error {
	defer helper.ExecutionTime(time.Now(), "[Storage:IconByHash]")

	err := s.db.QueryRow(`SELECT id FROM icons WHERE hash=$1`, icon.Hash).Scan(&icon.ID)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return fmt.Errorf("Unable to fetch icon by hash: %v", err)
	}

	return nil
}

// CreateIcon creates a new icon.
func (s *Storage) CreateIcon(icon *model.Icon) error {
	defer helper.ExecutionTime(time.Now(), "[Storage:CreateIcon]")

	query := `
		INSERT INTO icons
		(hash, mime_type, content)
		VALUES
		($1, $2, $3)
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		icon.Hash,
		normalizeMimeType(icon.MimeType),
		icon.Content,
	).Scan(&icon.ID)

	if err != nil {
		return fmt.Errorf("Unable to create icon: %v", err)
	}

	return nil
}

// CreateFeedIcon creates an icon and associate the icon to the given feed.
func (s *Storage) CreateFeedIcon(feed *model.Feed, icon *model.Icon) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:CreateFeedIcon] feedID=%d", feed.ID))

	err := s.IconByHash(icon)
	if err != nil {
		return err
	}

	if icon.ID == 0 {
		err := s.CreateIcon(icon)
		if err != nil {
			return err
		}
	}

	_, err = s.db.Exec(`INSERT INTO feed_icons (feed_id, icon_id) VALUES ($1, $2)`, feed.ID, icon.ID)
	if err != nil {
		return fmt.Errorf("unable to create feed icon: %v", err)
	}

	return nil
}

// Icons returns all icons tht belongs to a user.
func (s *Storage) Icons(userID int64) (model.Icons, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:Icons] userID=%d", userID))
	query := `
		SELECT
		icons.id, icons.hash, icons.mime_type, icons.content
		FROM icons
		LEFT JOIN feed_icons ON feed_icons.icon_id=icons.id
		LEFT JOIN feeds ON feeds.id=feed_icons.feed_id
		WHERE feeds.user_id=$1
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch icons: %v", err)
	}
	defer rows.Close()

	var icons model.Icons
	for rows.Next() {
		var icon model.Icon
		err := rows.Scan(&icon.ID, &icon.Hash, &icon.MimeType, &icon.Content)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch icons row: %v", err)
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
