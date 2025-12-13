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
	query := `SELECT true FROM integrations WHERE user_id != $1 AND fever_username=$2 LIMIT 1`
	var result bool
	s.db.QueryRow(query, userID, feverUsername).Scan(&result)
	return result
}

// HasDuplicateGoogleReaderUsername checks if another user have the same Google Reader username.
func (s *Storage) HasDuplicateGoogleReaderUsername(userID int64, googleReaderUsername string) bool {
	query := `SELECT true FROM integrations WHERE user_id != $1 AND googlereader_username=$2 LIMIT 1`
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
			wallabag_tags,
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
			linkwarden_collection_id,
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
			readeck_push_enabled,
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
			karakeep_url,
			karakeep_tags,
			linktaco_enabled,
			linktaco_api_token,
			linktaco_org_slug,
			linktaco_tags,
			linktaco_visibility,
			archiveorg_enabled
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
		&integration.WallabagTags,
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
		&integration.LinkwardenCollectionID,
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
		&integration.ReadeckPushEnabled,
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
		&integration.KarakeepTags,
		&integration.LinktacoEnabled,
		&integration.LinktacoAPIToken,
		&integration.LinktacoOrgSlug,
		&integration.LinktacoTags,
		&integration.LinktacoVisibility,
		&integration.ArchiveorgEnabled,
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
			wallabag_tags=$18,
			nunux_keeper_enabled=$19,
			nunux_keeper_url=$20,
			nunux_keeper_api_key=$21,
			googlereader_enabled=$22,
			googlereader_username=$23,
			googlereader_password=$24,
			telegram_bot_enabled=$25,
			telegram_bot_token=$26,
			telegram_bot_chat_id=$27,
			telegram_bot_topic_id=$28,
			telegram_bot_disable_web_page_preview=$29,
			telegram_bot_disable_notification=$30,
			telegram_bot_disable_buttons=$31,
			espial_enabled=$32,
			espial_url=$33,
			espial_api_key=$34,
			espial_tags=$35,
			linkace_enabled=$36,
			linkace_url=$37,
			linkace_api_key=$38,
			linkace_tags=$39,
			linkace_is_private=$40,
			linkace_check_disabled=$41,
			linkding_enabled=$42,
			linkding_url=$43,
			linkding_api_key=$44,
			linkding_tags=$45,
			linkding_mark_as_unread=$46,
			matrix_bot_enabled=$47,
			matrix_bot_user=$48,
			matrix_bot_password=$49,
			matrix_bot_url=$50,
			matrix_bot_chat_id=$51,
			notion_enabled=$52,
			notion_token=$53,
			notion_page_id=$54,
			readwise_enabled=$55,
			readwise_api_key=$56,
			apprise_enabled=$57,
			apprise_url=$58,
			apprise_services_url=$59,
			readeck_enabled=$60,
			readeck_url=$61,
			readeck_api_key=$62,
			readeck_labels=$63,
			readeck_only_url=$64,
			shiori_enabled=$65,
			shiori_url=$66,
			shiori_username=$67,
			shiori_password=$68,
			shaarli_enabled=$69,
			shaarli_url=$70,
			shaarli_api_secret=$71,
			webhook_enabled=$72,
			webhook_url=$73,
			webhook_secret=$74,
			rssbridge_enabled=$75,
			rssbridge_url=$76,
			omnivore_enabled=$77,
			omnivore_api_key=$78,
			omnivore_url=$79,
			linkwarden_enabled=$80,
			linkwarden_url=$81,
			linkwarden_api_key=$82,
			raindrop_enabled=$83,
			raindrop_token=$84,
			raindrop_collection_id=$85,
			raindrop_tags=$86,
			betula_enabled=$87,
			betula_url=$88,
			betula_token=$89,
			ntfy_enabled=$90,
			ntfy_topic=$91,
			ntfy_url=$92,
			ntfy_api_token=$93,
			ntfy_username=$94,
			ntfy_password=$95,
			ntfy_icon_url=$96,
			ntfy_internal_links=$97,
			cubox_enabled=$98,
			cubox_api_link=$99,
			discord_enabled=$100,
			discord_webhook_link=$101,
			slack_enabled=$102,
			slack_webhook_link=$103,
			pushover_enabled=$104,
			pushover_user=$105,
			pushover_token=$106,
			pushover_device=$107,
			pushover_prefix=$108,
			rssbridge_token=$109,
			karakeep_enabled=$110,
			karakeep_api_key=$111,
			karakeep_url=$112,
			karakeep_tags=$113,
			linktaco_enabled=$114,
			linktaco_api_token=$115,
			linktaco_org_slug=$116,
			linktaco_tags=$117,
			linktaco_visibility=$118,
			archiveorg_enabled=$119,
			linkwarden_collection_id=$120,
			readeck_push_enabled=$121
		WHERE
			user_id=$122
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
		integration.WallabagTags,
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
		integration.KarakeepTags,
		integration.LinktacoEnabled,
		integration.LinktacoAPIToken,
		integration.LinktacoOrgSlug,
		integration.LinktacoTags,
		integration.LinktacoVisibility,
		integration.ArchiveorgEnabled,
		integration.LinkwardenCollectionID,
		integration.ReadeckPushEnabled,
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
				linktaco_enabled='t' OR
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
				slack_enabled='t' OR
				archiveorg_enabled='t'
			)
	`
	if err := s.db.QueryRow(query, userID).Scan(&result); err != nil {
		result = false
	}

	return result
}
