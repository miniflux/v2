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
			omnivore_url,
			raindrop_enabled,
			raindrop_token,
			raindrop_collection_id,
			raindrop_tags,
			betula_enabled,
			betula_url,
			betula_token,
			ntfy_enabled,
			ntfy_topic,
			ntfy_url,
			ntfy_api_token,
			ntfy_username,
			ntfy_password,
			ntfy_icon_url,
			ntfy_internal_links,
			cubox_enabled,
			cubox_api_link,
			discord_enabled,
			discord_webhook_link,
			slack_enabled,
			slack_webhook_link,
			pushover_enabled,
			pushover_user,
			pushover_token,
			pushover_device,
			pushover_prefix,
			rssbridge_token,
			karakeep_enabled,
			karakeep_api_key,
			karakeep_url
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
		&integration.RaindropEnabled,
		&integration.RaindropToken,
		&integration.RaindropCollectionID,
		&integration.RaindropTags,
		&integration.BetulaEnabled,
		&integration.BetulaURL,
		&integration.BetulaToken,
		&integration.NtfyEnabled,
		&integration.NtfyTopic,
		&integration.NtfyURL,
		&integration.NtfyAPIToken,
		&integration.NtfyUsername,
		&integration.NtfyPassword,
		&integration.NtfyIconURL,
		&integration.NtfyInternalLinks,
		&integration.CuboxEnabled,
		&integration.CuboxAPILink,
		&integration.DiscordEnabled,
		&integration.DiscordWebhookLink,
		&integration.SlackEnabled,
		&integration.SlackWebhookLink,
		&integration.PushoverEnabled,
		&integration.PushoverUser,
		&integration.PushoverToken,
		&integration.PushoverDevice,
		&integration.PushoverPrefix,
		&integration.RSSBridgeToken,
		&integration.KarakeepEnabled,
		&integration.KarakeepAPIKey,
		&integration.KarakeepURL,
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
			googlereader_enabled=$21,
			googlereader_username=$22,
			googlereader_password=$23,
			telegram_bot_enabled=$24,
			telegram_bot_token=$25,
			telegram_bot_chat_id=$26,
			telegram_bot_topic_id=$27,
			telegram_bot_disable_web_page_preview=$28,
			telegram_bot_disable_notification=$29,
			telegram_bot_disable_buttons=$30,
			espial_enabled=$31,
			espial_url=$32,
			espial_api_key=$33,
			espial_tags=$34,
			linkace_enabled=$35,
			linkace_url=$36,
			linkace_api_key=$37,
			linkace_tags=$38,
			linkace_is_private=$39,
			linkace_check_disabled=$40,
			linkding_enabled=$41,
			linkding_url=$42,
			linkding_api_key=$43,
			linkding_tags=$44,
			linkding_mark_as_unread=$45,
			matrix_bot_enabled=$46,
			matrix_bot_user=$47,
			matrix_bot_password=$48,
			matrix_bot_url=$49,
			matrix_bot_chat_id=$50,
			notion_enabled=$51,
			notion_token=$52,
			notion_page_id=$53,
			readwise_enabled=$54,
			readwise_api_key=$55,
			apprise_enabled=$56,
			apprise_url=$57,
			apprise_services_url=$58,
			readeck_enabled=$59,
			readeck_url=$60,
			readeck_api_key=$61,
			readeck_labels=$62,
			readeck_only_url=$63,
			shiori_enabled=$64,
			shiori_url=$65,
			shiori_username=$66,
			shiori_password=$67,
			shaarli_enabled=$68,
			shaarli_url=$69,
			shaarli_api_secret=$70,
			webhook_enabled=$71,
			webhook_url=$72,
			webhook_secret=$73,
			rssbridge_enabled=$74,
			rssbridge_url=$75,
			omnivore_enabled=$76,
			omnivore_api_key=$77,
			omnivore_url=$78,
			linkwarden_enabled=$79,
			linkwarden_url=$80,
			linkwarden_api_key=$81,
			raindrop_enabled=$82,
			raindrop_token=$83,
			raindrop_collection_id=$84,
			raindrop_tags=$85,
			betula_enabled=$86,
			betula_url=$87,
			betula_token=$88,
			ntfy_enabled=$89,
			ntfy_topic=$90,
			ntfy_url=$91,
			ntfy_api_token=$92,
			ntfy_username=$93,
			ntfy_password=$94,
			ntfy_icon_url=$95,
			ntfy_internal_links=$96,
			cubox_enabled=$97,
			cubox_api_link=$98,
			discord_enabled=$99,
			discord_webhook_link=$100,
			slack_enabled=$101,
			slack_webhook_link=$102,
			pushover_enabled=$103,
			pushover_user=$104,
			pushover_token=$105,
			pushover_device=$106,
			pushover_prefix=$107,
			rssbridge_token=$108,
			karakeep_enabled=$109,
			karakeep_api_key=$110,
			karakeep_url=$111
		WHERE
			user_id=$112
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
		integration.RaindropEnabled,
		integration.RaindropToken,
		integration.RaindropCollectionID,
		integration.RaindropTags,
		integration.BetulaEnabled,
		integration.BetulaURL,
		integration.BetulaToken,
		integration.NtfyEnabled,
		integration.NtfyTopic,
		integration.NtfyURL,
		integration.NtfyAPIToken,
		integration.NtfyUsername,
		integration.NtfyPassword,
		integration.NtfyIconURL,
		integration.NtfyInternalLinks,
		integration.CuboxEnabled,
		integration.CuboxAPILink,
		integration.DiscordEnabled,
		integration.DiscordWebhookLink,
		integration.SlackEnabled,
		integration.SlackWebhookLink,
		integration.PushoverEnabled,
		integration.PushoverUser,
		integration.PushoverToken,
		integration.PushoverDevice,
		integration.PushoverPrefix,
		integration.RSSBridgeToken,
		integration.KarakeepEnabled,
		integration.KarakeepAPIKey,
		integration.KarakeepURL,
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
				linkace_enabled='t' OR
				linkding_enabled='t' OR
				linkwarden_enabled='t' OR
				apprise_enabled='t' OR
				shiori_enabled='t' OR
				readeck_enabled='t' OR
				shaarli_enabled='t' OR
				webhook_enabled='t' OR
				omnivore_enabled='t' OR
				karakeep_enabled='t' OR
				raindrop_enabled='t' OR
				betula_enabled='t' OR
				cubox_enabled='t' OR
				discord_enabled='t' OR
				slack_enabled='t'
			)
	`
	if err := s.db.QueryRow(query, userID).Scan(&result); err != nil {
		result = false
	}

	return result
}
