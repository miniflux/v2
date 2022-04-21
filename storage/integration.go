// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"miniflux.app/model"
)

// HasDuplicateFeverUsername checks if another user have the same Fever username.
func (s *Storage) HasDuplicateFeverUsername(userID int64, feverUsername string) bool {
	query := `SELECT true FROM integrations WHERE user_id != $1 AND fever_username=$2`
	var result bool
	s.db.QueryRow(query, userID, feverUsername).Scan(&result)
	return result
}

// HasDuplicateGoogleReaderUsername checks if another user have the same Google Reader username.
func (s *Storage) HasDuplicateGoogleReaderUsername(userID int64, googleReaderUsername string) bool {
	query := `SELECT true FROM integrations WHERE user_id != $1 AND googlereader_username=$2`
	var result bool
	s.db.QueryRow(query, userID, googleReaderUsername).Scan(&result)
	return result
}

// UserByFeverToken returns a user by using the Fever API token.
func (s *Storage) UserByFeverToken(token string) (*model.User, error) {
	query := `
		SELECT
			users.id, users.is_admin, users.timezone
		FROM
			users
		LEFT JOIN
			integrations ON integrations.user_id=users.id
		WHERE
			integrations.fever_enabled='t' AND lower(integrations.fever_token)=lower($1)
	`

	var user model.User
	err := s.db.QueryRow(query, token).Scan(&user.ID, &user.IsAdmin, &user.Timezone)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("store: unable to fetch user: %v", err)
	default:
		return &user, nil
	}
}

// GoogleReaderUserCheckPassword validates the Google Reader hashed password.
func (s *Storage) GoogleReaderUserCheckPassword(username, password string) error {
	var hash string

	query := `
		SELECT
			googlereader_password
		FROM
			integrations
		WHERE
			integrations.googlereader_enabled='t' AND integrations.googlereader_username=$1
	`

	err := s.db.QueryRow(query, username).Scan(&hash)
	if err == sql.ErrNoRows {
		return fmt.Errorf(`store: unable to find this user: %s`, username)
	} else if err != nil {
		return fmt.Errorf(`store: unable to fetch user: %v`, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf(`store: invalid password for "%s" (%v)`, username, err)
	}

	return nil
}

// GoogleReaderUserGetIntegration returns part of the Google Reader parts of the integration struct.
func (s *Storage) GoogleReaderUserGetIntegration(username string) (*model.Integration, error) {
	var integration model.Integration

	query := `
		SELECT
			user_id,
			googlereader_enabled,
			googlereader_username,
			googlereader_password
		FROM
			integrations
		WHERE
			integrations.googlereader_enabled='t' AND integrations.googlereader_username=$1
	`

	err := s.db.QueryRow(query, username).Scan(&integration.UserID, &integration.GoogleReaderEnabled, &integration.GoogleReaderUsername, &integration.GoogleReaderPassword)
	if err == sql.ErrNoRows {
		return &integration, fmt.Errorf(`store: unable to find this user: %s`, username)
	} else if err != nil {
		return &integration, fmt.Errorf(`store: unable to fetch user: %v`, err)
	}

	return &integration, nil
}

// Integration returns user integration settings.
func (s *Storage) Integration(userID int64) (*model.Integration, error) {
	query := `
		SELECT
			user_id,
			pinboard_enabled,
			pinboard_token,
			pinboard_tags,
			pinboard_mark_as_unread,
			instapaper_enabled,
			instapaper_username,
			instapaper_password,
			fever_enabled,
			fever_username,
			fever_token,
			googlereader_enabled,
			googlereader_username,
			googlereader_password,
			wallabag_enabled,
			wallabag_url,
			wallabag_client_id,
			wallabag_client_secret,
			wallabag_username,
			wallabag_password,
			nunux_keeper_enabled,
			nunux_keeper_url,
			nunux_keeper_api_key,
			espial_enabled,
			espial_url,
			espial_api_key,
			espial_tags,
			pocket_enabled,
			pocket_access_token,
			pocket_consumer_key,
			telegram_bot_enabled,
			telegram_bot_token,
			telegram_bot_chat_id
		FROM
			integrations
		WHERE
			user_id=$1
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
		&integration.FeverEnabled,
		&integration.FeverUsername,
		&integration.FeverToken,
		&integration.GoogleReaderEnabled,
		&integration.GoogleReaderUsername,
		&integration.GoogleReaderPassword,
		&integration.WallabagEnabled,
		&integration.WallabagURL,
		&integration.WallabagClientID,
		&integration.WallabagClientSecret,
		&integration.WallabagUsername,
		&integration.WallabagPassword,
		&integration.NunuxKeeperEnabled,
		&integration.NunuxKeeperURL,
		&integration.NunuxKeeperAPIKey,
		&integration.EspialEnabled,
		&integration.EspialURL,
		&integration.EspialAPIKey,
		&integration.EspialTags,
		&integration.PocketEnabled,
		&integration.PocketAccessToken,
		&integration.PocketConsumerKey,
		&integration.TelegramBotEnabled,
		&integration.TelegramBotToken,
		&integration.TelegramBotChatID,
	)
	switch {
	case err == sql.ErrNoRows:
		return &integration, nil
	case err != nil:
		return &integration, fmt.Errorf(`store: unable to fetch integration row: %v`, err)
	default:
		return &integration, nil
	}
}

// UpdateIntegration saves user integration settings.
func (s *Storage) UpdateIntegration(integration *model.Integration) error {
	var err error
	if integration.GoogleReaderPassword != "" {
		integration.GoogleReaderPassword, err = hashPassword(integration.GoogleReaderPassword)
		if err != nil {
			return err
		}
		query := `
		UPDATE
			integrations
		SET
			pinboard_enabled=$1,
			pinboard_token=$2,
			pinboard_tags=$3,
			pinboard_mark_as_unread=$4,
			instapaper_enabled=$5,
			instapaper_username=$6,
			instapaper_password=$7,
			fever_enabled=$8,
			fever_username=$9,
			fever_token=$10,
			wallabag_enabled=$11,
			wallabag_url=$12,
			wallabag_client_id=$13,
			wallabag_client_secret=$14,
			wallabag_username=$15,
			wallabag_password=$16,
			nunux_keeper_enabled=$17,
			nunux_keeper_url=$18,
			nunux_keeper_api_key=$19,
			pocket_enabled=$20,
			pocket_access_token=$21,
			pocket_consumer_key=$22,
			googlereader_enabled=$23,
			googlereader_username=$24,
			googlereader_password=$25,
			telegram_bot_enabled=$26,
			telegram_bot_token=$27,
			telegram_bot_chat_id=$28,
			espial_enabled=$29,
			espial_url=$30,
			espial_api_key=$31,
			espial_tags=$32,
		WHERE
			user_id=$33
	`
		_, err = s.db.Exec(
			query,
			integration.PinboardEnabled,
			integration.PinboardToken,
			integration.PinboardTags,
			integration.PinboardMarkAsUnread,
			integration.InstapaperEnabled,
			integration.InstapaperUsername,
			integration.InstapaperPassword,
			integration.FeverEnabled,
			integration.FeverUsername,
			integration.FeverToken,
			integration.WallabagEnabled,
			integration.WallabagURL,
			integration.WallabagClientID,
			integration.WallabagClientSecret,
			integration.WallabagUsername,
			integration.WallabagPassword,
			integration.NunuxKeeperEnabled,
			integration.NunuxKeeperURL,
			integration.NunuxKeeperAPIKey,
			integration.PocketEnabled,
			integration.PocketAccessToken,
			integration.PocketConsumerKey,
			integration.GoogleReaderEnabled,
			integration.GoogleReaderUsername,
			integration.GoogleReaderPassword,
			integration.TelegramBotEnabled,
			integration.TelegramBotToken,
			integration.TelegramBotChatID,
			integration.EspialEnabled,
			integration.EspialURL,
			integration.EspialAPIKey,
			integration.EspialTags,
			integration.UserID,
		)
	} else {
		query := `
		UPDATE
			integrations
		SET
			pinboard_enabled=$1,
			pinboard_token=$2,
			pinboard_tags=$3,
			pinboard_mark_as_unread=$4,
			instapaper_enabled=$5,
			instapaper_username=$6,
			instapaper_password=$7,
			fever_enabled=$8,
			fever_username=$9,
			fever_token=$10,
			wallabag_enabled=$11,
			wallabag_url=$12,
			wallabag_client_id=$13,
			wallabag_client_secret=$14,
			wallabag_username=$15,
			wallabag_password=$16,
			nunux_keeper_enabled=$17,
			nunux_keeper_url=$18,
			nunux_keeper_api_key=$19,
			pocket_enabled=$20,
			pocket_access_token=$21,
			pocket_consumer_key=$22,
			googlereader_enabled=$23,
			googlereader_username=$24,
		    googlereader_password=$25,
			telegram_bot_enabled=$26,
			telegram_bot_token=$27,
			telegram_bot_chat_id=$28,
			espial_enabled=$29,
			espial_url=$30,
			espial_api_key=$31,
			espial_tags=$32
		WHERE
			user_id=$33
	`
		_, err = s.db.Exec(
			query,
			integration.PinboardEnabled,
			integration.PinboardToken,
			integration.PinboardTags,
			integration.PinboardMarkAsUnread,
			integration.InstapaperEnabled,
			integration.InstapaperUsername,
			integration.InstapaperPassword,
			integration.FeverEnabled,
			integration.FeverUsername,
			integration.FeverToken,
			integration.WallabagEnabled,
			integration.WallabagURL,
			integration.WallabagClientID,
			integration.WallabagClientSecret,
			integration.WallabagUsername,
			integration.WallabagPassword,
			integration.NunuxKeeperEnabled,
			integration.NunuxKeeperURL,
			integration.NunuxKeeperAPIKey,
			integration.PocketEnabled,
			integration.PocketAccessToken,
			integration.PocketConsumerKey,
			integration.GoogleReaderEnabled,
			integration.GoogleReaderUsername,
			integration.GoogleReaderPassword,
			integration.TelegramBotEnabled,
			integration.TelegramBotToken,
			integration.TelegramBotChatID,
			integration.EspialEnabled,
			integration.EspialURL,
			integration.EspialAPIKey,
			integration.EspialTags,
			integration.UserID,
		)
	}

	if err != nil {
		return fmt.Errorf(`store: unable to update integration row: %v`, err)
	}

	return nil
}

// HasSaveEntry returns true if the given user can save articles to third-parties.
func (s *Storage) HasSaveEntry(userID int64) (result bool) {
	query := `
		SELECT
			true
		FROM
			integrations
		WHERE
			user_id=$1
		AND
			(pinboard_enabled='t' OR instapaper_enabled='t' OR wallabag_enabled='t' OR nunux_keeper_enabled='t' OR espial_enabled='t' OR pocket_enabled='t')
	`
	if err := s.db.QueryRow(query, userID).Scan(&result); err != nil {
		result = false
	}

	return result
}
