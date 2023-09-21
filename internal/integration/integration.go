// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package integration // import "miniflux.app/v2/internal/integration"

import (
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/integration/apprise"
	"miniflux.app/v2/internal/integration/espial"
	"miniflux.app/v2/internal/integration/instapaper"
	"miniflux.app/v2/internal/integration/linkding"
	"miniflux.app/v2/internal/integration/matrixbot"
	"miniflux.app/v2/internal/integration/notion"
	"miniflux.app/v2/internal/integration/nunuxkeeper"
	"miniflux.app/v2/internal/integration/pinboard"
	"miniflux.app/v2/internal/integration/pocket"
	"miniflux.app/v2/internal/integration/readwise"
	"miniflux.app/v2/internal/integration/shaarli"
	"miniflux.app/v2/internal/integration/shiori"
	"miniflux.app/v2/internal/integration/telegrambot"
	"miniflux.app/v2/internal/integration/wallabag"
	"miniflux.app/v2/internal/integration/webhook"
	"miniflux.app/v2/internal/logger"
	"miniflux.app/v2/internal/model"
)

// SendEntry sends the entry to third-party providers when the user click on "Save".
func SendEntry(entry *model.Entry, integration *model.Integration) {
	if integration.PinboardEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Pinboard", entry.ID, entry.URL, integration.UserID)

		client := pinboard.NewClient(integration.PinboardToken)
		err := client.CreateBookmark(
			entry.URL,
			entry.Title,
			integration.PinboardTags,
			integration.PinboardMarkAsUnread,
		)

		if err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.InstapaperEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Instapaper", entry.ID, entry.URL, integration.UserID)

		client := instapaper.NewClient(integration.InstapaperUsername, integration.InstapaperPassword)
		if err := client.AddURL(entry.URL, entry.Title); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.WallabagEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Wallabag", entry.ID, entry.URL, integration.UserID)

		client := wallabag.NewClient(
			integration.WallabagURL,
			integration.WallabagClientID,
			integration.WallabagClientSecret,
			integration.WallabagUsername,
			integration.WallabagPassword,
			integration.WallabagOnlyURL,
		)

		if err := client.CreateEntry(entry.URL, entry.Title, entry.Content); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.NotionEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Notion", entry.ID, entry.URL, integration.UserID)

		client := notion.NewClient(
			integration.NotionToken,
			integration.NotionPageID,
		)
		if err := client.UpdateDocument(entry.URL, entry.Title); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.NunuxKeeperEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to NunuxKeeper", entry.ID, entry.URL, integration.UserID)

		client := nunuxkeeper.NewClient(
			integration.NunuxKeeperURL,
			integration.NunuxKeeperAPIKey,
		)

		if err := client.AddEntry(entry.URL, entry.Title, entry.Content); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.EspialEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Espial", entry.ID, entry.URL, integration.UserID)

		client := espial.NewClient(
			integration.EspialURL,
			integration.EspialAPIKey,
		)

		if err := client.CreateLink(entry.URL, entry.Title, integration.EspialTags); err != nil {
			logger.Error("[Integration] Unable to send entry #%d to Espial for user #%d: %v", entry.ID, integration.UserID, err)
		}
	}

	if integration.PocketEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Pocket", entry.ID, entry.URL, integration.UserID)

		client := pocket.NewClient(config.Opts.PocketConsumerKey(integration.PocketConsumerKey), integration.PocketAccessToken)
		if err := client.AddURL(entry.URL, entry.Title); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.LinkdingEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Linkding", entry.ID, entry.URL, integration.UserID)

		client := linkding.NewClient(
			integration.LinkdingURL,
			integration.LinkdingAPIKey,
			integration.LinkdingTags,
			integration.LinkdingMarkAsUnread,
		)
		if err := client.CreateBookmark(entry.URL, entry.Title); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.ReadwiseEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Readwise Reader", entry.ID, entry.URL, integration.UserID)

		client := readwise.NewClient(
			integration.ReadwiseAPIKey,
		)

		if err := client.CreateDocument(entry.URL); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.ShioriEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Shiori", entry.ID, entry.URL, integration.UserID)

		client := shiori.NewClient(
			integration.ShioriURL,
			integration.ShioriUsername,
			integration.ShioriPassword,
		)

		if err := client.CreateBookmark(entry.URL, entry.Title); err != nil {
			logger.Error("[Integration] Unable to send entry #%d to Shiori for user #%d: %v", entry.ID, integration.UserID, err)
		}
	}

	if integration.ShaarliEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Shaarli", entry.ID, entry.URL, integration.UserID)

		client := shaarli.NewClient(
			integration.ShaarliURL,
			integration.ShaarliAPISecret,
		)

		if err := client.CreateLink(entry.URL, entry.Title); err != nil {
			logger.Error("[Integration] Unable to send entry #%d to Shaarli for user #%d: %v", entry.ID, integration.UserID, err)
		}
	}

	if integration.WebhookEnabled {
		logger.Debug("[Integration] Sending entry #%d %q for user #%d to Webhook URL: %s", entry.ID, entry.URL, integration.UserID, integration.WebhookURL)

		webhookClient := webhook.NewClient(integration.WebhookURL, integration.WebhookSecret)
		if err := webhookClient.SendSaveEntryWebhookEvent(entry); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}
}

// PushEntries pushes a list of entries to activated third-party providers during feed refreshes.
func PushEntries(feed *model.Feed, entries model.Entries, userIntegrations *model.Integration) {
	if userIntegrations.MatrixBotEnabled {
		logger.Debug("[Integration] Sending %d entries for user #%d to Matrix", len(entries), userIntegrations.UserID)

		err := matrixbot.PushEntries(feed, entries, userIntegrations.MatrixBotURL, userIntegrations.MatrixBotUser, userIntegrations.MatrixBotPassword, userIntegrations.MatrixBotChatID)
		if err != nil {
			logger.Error("[Integration] push entries to matrix bot failed: %v", err)
		}
	}

	if userIntegrations.WebhookEnabled {
		logger.Debug("[Integration] Sending %d entries for user #%d to Webhook URL: %s", len(entries), userIntegrations.UserID, userIntegrations.WebhookURL)

		webhookClient := webhook.NewClient(userIntegrations.WebhookURL, userIntegrations.WebhookSecret)
		if err := webhookClient.SendNewEntriesWebhookEvent(feed, entries); err != nil {
			logger.Error("[Integration] sending entries to webhook failed: %v", err)
		}
	}

	// Integrations that only support sending individual entries
	if userIntegrations.TelegramBotEnabled || userIntegrations.AppriseEnabled {
		for _, entry := range entries {
			if userIntegrations.TelegramBotEnabled {
				logger.Debug("[Integration] Sending entry %q for user #%d to Telegram", entry.URL, userIntegrations.UserID)

				if err := telegrambot.PushEntry(
					feed,
					entry,
					userIntegrations.TelegramBotToken,
					userIntegrations.TelegramBotChatID,
					userIntegrations.TelegramBotTopicID,
					userIntegrations.TelegramBotDisableWebPagePreview,
					userIntegrations.TelegramBotDisableNotification,
					userIntegrations.TelegramBotDisableButtons,
				); err != nil {
					logger.Error("[Integration] %v", err)
				}
			}

			if userIntegrations.AppriseEnabled {
				logger.Debug("[Integration] Sending entry %q for user #%d to Apprise", entry.URL, userIntegrations.UserID)

				appriseServiceURLs := userIntegrations.AppriseURL
				if feed.AppriseServiceURLs != "" {
					appriseServiceURLs = feed.AppriseServiceURLs
				}

				client := apprise.NewClient(
					userIntegrations.AppriseServicesURL,
					appriseServiceURLs,
				)

				if err := client.SendNotification(entry); err != nil {
					logger.Error("[Integration] %v", err)
				}
			}
		}
	}
}
