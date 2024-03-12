// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"miniflux.app/v2/internal/model"
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
			users.id, users.username, users.is_admin, users.timezone
		FROM
			users
		LEFT JOIN
			integrations ON integrations.user_id=users.id
		WHERE
			integrations.fever_enabled='t' AND lower(integrations.fever_token)=lower($1)
	`

	var user model.User
	err := s.db.QueryRow(query, token).Scan(&user.ID, &user.Username, &user.IsAdmin, &user.Timezone)
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
			wallabag_only_url,
			wallabag_url,
			wallabag_client_id,
			wallabag_client_secret,
			wallabag_username,
			wallabag_password,
			notion_enabled,
			notion_token,
			notion_page_id,
			nunux_keeper_enabled,
			nunux_keeper_url,
			nunux_keeper_api_key,
			espial_enabled,
			espial_url,
			espial_api_key,
			espial_tags,
			readwise_enabled,
			readwise_api_key,
			pocket_enabled,
			pocket_access_token,
			pocket_consumer_key,
			telegram_bot_enabled,
			telegram_bot_token,
			telegram_bot_chat_id,
			telegram_bot_topic_id,
			telegram_bot_disable_web_page_preview,
			telegram_bot_disable_notification,
			telegram_bot_disable_buttons,
			linkace_enabled,
			linkace_url,
			linkace_api_key,
			linkace_tags,
			linkace_is_private,
			linkace_check_disabled,
			linkding_enabled,
			linkding_url,
			linkding_api_key,
			linkding_tags,
			linkding_mark_as_unread,
			linkwarden_enabled,
			linkwarden_url,
			linkwarden_api_key,
			matrix_bot_enabled,
			matrix_bot_user,
			matrix_bot_password,
			matrix_bot_url,
			matrix_bot_chat_id,
			apprise_enabled,
			apprise_url,
			apprise_services_url,
			readeck_enabled,
			readeck_url,
			readeck_api_key,
			readeck_labels,
			readeck_only_url,
			shiori_enabled,
			shiori_url,
			shiori_username,
			shiori_password,
			shaarli_enabled,
			shaarli_url,
			shaarli_api_secret,
			webhook_enabled,
			webhook_url,
			webhook_secret,
			rssbridge_enabled,
			rssbridge_url,
			omnivore_enabled,
			omnivore_api_key,
			omnivore_url
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
		&integration.WallabagOnlyURL,
		&integration.WallabagURL,
		&integration.WallabagClientID,
		&integration.WallabagClientSecret,
		&integration.WallabagUsername,
		&integration.WallabagPassword,
		&integration.NotionEnabled,
		&integration.NotionToken,
		&integration.NotionPageID,
		&integration.NunuxKeeperEnabled,
		&integration.NunuxKeeperURL,
		&integration.NunuxKeeperAPIKey,
		&integration.EspialEnabled,
		&integration.EspialURL,
		&integration.EspialAPIKey,
		&integration.EspialTags,
		&integration.ReadwiseEnabled,
		&integration.ReadwiseAPIKey,
		&integration.PocketEnabled,
		&integration.PocketAccessToken,
		&integration.PocketConsumerKey,
		&integration.TelegramBotEnabled,
		&integration.TelegramBotToken,
		&integration.TelegramBotChatID,
		&integration.TelegramBotTopicID,
		&integration.TelegramBotDisableWebPagePreview,
		&integration.TelegramBotDisableNotification,
		&integration.TelegramBotDisableButtons,
		&integration.LinkAceEnabled,
		&integration.LinkAceURL,
		&integration.LinkAceAPIKey,
		&integration.LinkAceTags,
		&integration.LinkAcePrivate,
		&integration.LinkAceCheckDisabled,
		&integration.LinkdingEnabled,
		&integration.LinkdingURL,
		&integration.LinkdingAPIKey,
		&integration.LinkdingTags,
		&integration.LinkdingMarkAsUnread,
		&integration.LinkwardenEnabled,
		&integration.LinkwardenURL,
		&integration.LinkwardenAPIKey,
		&integration.MatrixBotEnabled,
		&integration.MatrixBotUser,
		&integration.MatrixBotPassword,
		&integration.MatrixBotURL,
		&integration.MatrixBotChatID,
		&integration.AppriseEnabled,
		&integration.AppriseURL,
		&integration.AppriseServicesURL,
		&integration.ReadeckEnabled,
		&integration.ReadeckURL,
		&integration.ReadeckAPIKey,
		&integration.ReadeckLabels,
		&integration.ReadeckOnlyURL,
		&integration.ShioriEnabled,
		&integration.ShioriURL,
		&integration.ShioriUsername,
		&integration.ShioriPassword,
		&integration.ShaarliEnabled,
		&integration.ShaarliURL,
		&integration.ShaarliAPISecret,
		&integration.WebhookEnabled,
		&integration.WebhookURL,
		&integration.WebhookSecret,
		&integration.RSSBridgeEnabled,
		&integration.RSSBridgeURL,
		&integration.OmnivoreEnabled,
		&integration.OmnivoreAPIKey,
		&integration.OmnivoreURL,
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
			wallabag_only_url=$12,
			wallabag_url=$13,
			wallabag_client_id=$14,
			wallabag_client_secret=$15,
			wallabag_username=$16,
			wallabag_password=$17,
			nunux_keeper_enabled=$18,
			nunux_keeper_url=$19,
			nunux_keeper_api_key=$20,
			pocket_enabled=$21,
			pocket_access_token=$22,
			pocket_consumer_key=$23,
			googlereader_enabled=$24,
			googlereader_username=$25,
			googlereader_password=$26,
			telegram_bot_enabled=$27,
			telegram_bot_token=$28,
			telegram_bot_chat_id=$29,
			telegram_bot_topic_id=$30,
			telegram_bot_disable_web_page_preview=$31,
			telegram_bot_disable_notification=$32,
			telegram_bot_disable_buttons=$33,
			espial_enabled=$34,
			espial_url=$35,
			espial_api_key=$36,
			espial_tags=$37,
			linkace_enabled=$38,
			linkace_url=$39,
			linkace_api_key=$40,
			linkace_tags=$41,
			linkace_is_private=$42,
			linkace_check_disabled=$43,
			linkding_enabled=$44,
			linkding_url=$45,
			linkding_api_key=$46,
			linkding_tags=$47,
			linkding_mark_as_unread=$48,
			matrix_bot_enabled=$49,
			matrix_bot_user=$50,
			matrix_bot_password=$51,
			matrix_bot_url=$52,
			matrix_bot_chat_id=$53,
			notion_enabled=$54,
			notion_token=$55,
			notion_page_id=$56,
			readwise_enabled=$57,
			readwise_api_key=$58,
			apprise_enabled=$59,
			apprise_url=$60,
			apprise_services_url=$61,
			readeck_enabled=$62,
			readeck_url=$63,
			readeck_api_key=$64,
			readeck_labels=$65,
			readeck_only_url=$66,
			shiori_enabled=$67,
			shiori_url=$68,
			shiori_username=$69,
			shiori_password=$70,
			shaarli_enabled=$71,
			shaarli_url=$72,
			shaarli_api_secret=$73,
			webhook_enabled=$74,
			webhook_url=$75,
			webhook_secret=$76,
			rssbridge_enabled=$77,
			rssbridge_url=$78,
			omnivore_enabled=$79,
			omnivore_api_key=$80,
			omnivore_url=$81,
			linkwarden_enabled=$82,
			linkwarden_url=$83,
			linkwarden_api_key=$84
		WHERE
			user_id=$85
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
		integration.FeverEnabled,
		integration.FeverUsername,
		integration.FeverToken,
		integration.WallabagEnabled,
		integration.WallabagOnlyURL,
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
		integration.TelegramBotTopicID,
		integration.TelegramBotDisableWebPagePreview,
		integration.TelegramBotDisableNotification,
		integration.TelegramBotDisableButtons,
		integration.EspialEnabled,
		integration.EspialURL,
		integration.EspialAPIKey,
		integration.EspialTags,
		integration.LinkAceEnabled,
		integration.LinkAceURL,
		integration.LinkAceAPIKey,
		integration.LinkAceTags,
		integration.LinkAcePrivate,
		integration.LinkAceCheckDisabled,
		integration.LinkdingEnabled,
		integration.LinkdingURL,
		integration.LinkdingAPIKey,
		integration.LinkdingTags,
		integration.LinkdingMarkAsUnread,
		integration.MatrixBotEnabled,
		integration.MatrixBotUser,
		integration.MatrixBotPassword,
		integration.MatrixBotURL,
		integration.MatrixBotChatID,
		integration.NotionEnabled,
		integration.NotionToken,
		integration.NotionPageID,
		integration.ReadwiseEnabled,
		integration.ReadwiseAPIKey,
		integration.AppriseEnabled,
		integration.AppriseURL,
		integration.AppriseServicesURL,
		integration.ReadeckEnabled,
		integration.ReadeckURL,
		integration.ReadeckAPIKey,
		integration.ReadeckLabels,
		integration.ReadeckOnlyURL,
		integration.ShioriEnabled,
		integration.ShioriURL,
		integration.ShioriUsername,
		integration.ShioriPassword,
		integration.ShaarliEnabled,
		integration.ShaarliURL,
		integration.ShaarliAPISecret,
		integration.WebhookEnabled,
		integration.WebhookURL,
		integration.WebhookSecret,
		integration.RSSBridgeEnabled,
		integration.RSSBridgeURL,
		integration.OmnivoreEnabled,
		integration.OmnivoreAPIKey,
		integration.OmnivoreURL,
		integration.LinkwardenEnabled,
		integration.LinkwardenURL,
		integration.LinkwardenAPIKey,
		integration.UserID,
	)

	if err != nil {
		return fmt.Errorf(`store: unable to update integration record: %v`, err)
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
			(
				pinboard_enabled='t' OR
				instapaper_enabled='t' OR
				wallabag_enabled='t' OR
				notion_enabled='t' OR
				nunux_keeper_enabled='t' OR
				espial_enabled='t' OR
				readwise_enabled='t' OR
				pocket_enabled='t' OR
				linkace_enabled='t' OR
				linkding_enabled='t' OR
				linkwarden_enabled='t' OR
				apprise_enabled='t' OR
				shiori_enabled='t' OR
				readeck_enabled='t' OR
				shaarli_enabled='t' OR
				webhook_enabled='t' OR
				omnivore_enabled='t'
			)
	`
	if err := s.db.QueryRow(query, userID).Scan(&result); err != nil {
		result = false
	}

	return result
}
