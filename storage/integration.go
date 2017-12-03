// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"fmt"

	"github.com/miniflux/miniflux2/model"
)

// Integration returns user integration settings.
func (s *Storage) Integration(userID int64) (*model.Integration, error) {
	query := `SELECT
			user_id,
			pinboard_enabled,
			pinboard_token,
			pinboard_tags,
			pinboard_mark_as_unread,
			instapaper_enabled,
			instapaper_username,
			instapaper_password
		FROM integrations
		WHERE user_id=$1
	`
	var integration model.Integration
	err := s.db.QueryRow(query, userID).Scan(
		&integration.UserID,
		&integration.PinboardEnabled,
		&integration.PinboardToken,
		&integration.PinboardTags,
		&integration.PinboardMarkAsUnread,
		&integration.InstapaperEnabled,
		&integration.InstapaperUsername,
		&integration.InstapaperPassword,
	)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("unable to fetch integration row: %v", err)
	}

	return &integration, nil
}

// UpdateIntegration saves user integration settings.
func (s *Storage) UpdateIntegration(integration *model.Integration) error {
	query := `
		UPDATE integrations SET
			pinboard_enabled=$1,
			pinboard_token=$2,
			pinboard_tags=$3,
			pinboard_mark_as_unread=$4,
			instapaper_enabled=$5,
			instapaper_username=$6,
			instapaper_password=$7
		WHERE user_id=$8
	`
	_, err := s.db.Exec(
		query,
		integration.PinboardEnabled,
		integration.PinboardToken,
		integration.PinboardTags,
		integration.PinboardMarkAsUnread,
		integration.InstapaperEnabled,
		integration.InstapaperUsername,
		integration.InstapaperPassword,
		integration.UserID,
	)

	if err != nil {
		return fmt.Errorf("unable to update integration row: %v", err)
	}

	return nil
}

// CreateIntegration creates initial user integration settings.
func (s *Storage) CreateIntegration(userID int64) error {
	query := `INSERT INTO integrations (user_id) VALUES ($1)`
	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("unable to create integration row: %v", err)
	}

	return nil
}
