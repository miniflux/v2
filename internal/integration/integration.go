// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package integration // import "miniflux.app/v2/internal/integration"

import (
	"log/slog"

	"miniflux.app/v2/internal/integration/apprise"
	"miniflux.app/v2/internal/integration/betula"
	"miniflux.app/v2/internal/integration/cubox"
	"miniflux.app/v2/internal/integration/discord"
	"miniflux.app/v2/internal/integration/espial"
	"miniflux.app/v2/internal/integration/instapaper"
	"miniflux.app/v2/internal/integration/karakeep"
	"miniflux.app/v2/internal/integration/linkace"
	"miniflux.app/v2/internal/integration/linkding"
	"miniflux.app/v2/internal/integration/linkwarden"
	"miniflux.app/v2/internal/integration/matrixbot"
	"miniflux.app/v2/internal/integration/notion"
	"miniflux.app/v2/internal/integration/ntfy"
	"miniflux.app/v2/internal/integration/nunuxkeeper"
	"miniflux.app/v2/internal/integration/omnivore"
	"miniflux.app/v2/internal/integration/pinboard"
	"miniflux.app/v2/internal/integration/pushover"
	"miniflux.app/v2/internal/integration/raindrop"
	"miniflux.app/v2/internal/integration/readeck"
	"miniflux.app/v2/internal/integration/readwise"
	"miniflux.app/v2/internal/integration/shaarli"
	"miniflux.app/v2/internal/integration/shiori"
	"miniflux.app/v2/internal/integration/slack"
	"miniflux.app/v2/internal/integration/telegrambot"
	"miniflux.app/v2/internal/integration/wallabag"
	"miniflux.app/v2/internal/integration/webhook"
	"miniflux.app/v2/internal/model"
)

// SendEntry sends the entry to third-party providers when the user click on "Save".
func SendEntry(entry *model.Entry, userIntegrations *model.Integration) {
	if userIntegrations.BetulaEnabled {
		slog.Debug("Sending entry to Betula",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := betula.NewClient(userIntegrations.BetulaURL, userIntegrations.BetulaToken)
		err := client.CreateBookmark(
			entry.URL,
			entry.Title,
			entry.Tags,
		)

		if err != nil {
			slog.Error("Unable to send entry to Betula",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.PinboardEnabled {
		slog.Debug("Sending entry to Pinboard",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := pinboard.NewClient(userIntegrations.PinboardToken)
		err := client.CreateBookmark(
			entry.URL,
			entry.Title,
			userIntegrations.PinboardTags,
			userIntegrations.PinboardMarkAsUnread,
		)

		if err != nil {
			slog.Error("Unable to send entry to Pinboard",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.InstapaperEnabled {
		slog.Debug("Sending entry to Instapaper",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := instapaper.NewClient(userIntegrations.InstapaperUsername, userIntegrations.InstapaperPassword)
		if err := client.AddURL(entry.URL, entry.Title); err != nil {
			slog.Error("Unable to send entry to Instapaper",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.WallabagEnabled {
		slog.Debug("Sending entry to Wallabag",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := wallabag.NewClient(
			userIntegrations.WallabagURL,
			userIntegrations.WallabagClientID,
			userIntegrations.WallabagClientSecret,
			userIntegrations.WallabagUsername,
			userIntegrations.WallabagPassword,
			userIntegrations.WallabagOnlyURL,
		)

		if err := client.CreateEntry(entry.URL, entry.Title, entry.Content); err != nil {
			slog.Error("Unable to send entry to Wallabag",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.NotionEnabled {
		slog.Debug("Sending entry to Notion",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := notion.NewClient(
			userIntegrations.NotionToken,
			userIntegrations.NotionPageID,
		)
		if err := client.UpdateDocument(entry.URL, entry.Title); err != nil {
			slog.Error("Unable to send entry to Notion",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.NunuxKeeperEnabled {
		slog.Debug("Sending entry to NunuxKeeper",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := nunuxkeeper.NewClient(
			userIntegrations.NunuxKeeperURL,
			userIntegrations.NunuxKeeperAPIKey,
		)

		if err := client.AddEntry(entry.URL, entry.Title, entry.Content); err != nil {
			slog.Error("Unable to send entry to NunuxKeeper",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.EspialEnabled {
		slog.Debug("Sending entry to Espial",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := espial.NewClient(
			userIntegrations.EspialURL,
			userIntegrations.EspialAPIKey,
		)

		if err := client.CreateLink(entry.URL, entry.Title, userIntegrations.EspialTags); err != nil {
			slog.Error("Unable to send entry to Espial",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.LinkAceEnabled {
		slog.Debug("Sending entry to LinkAce",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := linkace.NewClient(
			userIntegrations.LinkAceURL,
			userIntegrations.LinkAceAPIKey,
			userIntegrations.LinkAceTags,
			userIntegrations.LinkAcePrivate,
			userIntegrations.LinkAceCheckDisabled,
		)
		if err := client.AddURL(entry.URL, entry.Title); err != nil {
			slog.Error("Unable to send entry to LinkAce",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.LinkdingEnabled {
		slog.Debug("Sending entry to Linkding",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := linkding.NewClient(
			userIntegrations.LinkdingURL,
			userIntegrations.LinkdingAPIKey,
			userIntegrations.LinkdingTags,
			userIntegrations.LinkdingMarkAsUnread,
		)
		if err := client.CreateBookmark(entry.URL, entry.Title); err != nil {
			slog.Error("Unable to send entry to Linkding",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.LinkwardenEnabled {
		slog.Debug("Sending entry to linkwarden",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := linkwarden.NewClient(
			userIntegrations.LinkwardenURL,
			userIntegrations.LinkwardenAPIKey,
		)
		if err := client.CreateBookmark(entry.URL, entry.Title); err != nil {
			slog.Error("Unable to send entry to Linkwarden",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.ReadeckEnabled {
		slog.Debug("Sending entry to Readeck",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := readeck.NewClient(
			userIntegrations.ReadeckURL,
			userIntegrations.ReadeckAPIKey,
			userIntegrations.ReadeckLabels,
			userIntegrations.ReadeckOnlyURL,
		)
		if err := client.CreateBookmark(entry.URL, entry.Title, entry.Content); err != nil {
			slog.Error("Unable to send entry to Readeck",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.ReadwiseEnabled {
		slog.Debug("Sending entry to Readwise",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := readwise.NewClient(
			userIntegrations.ReadwiseAPIKey,
		)

		if err := client.CreateDocument(entry.URL); err != nil {
			slog.Error("Unable to send entry to Readwise",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.CuboxEnabled {
		slog.Debug("Sending entry to Cubox",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := cubox.NewClient(userIntegrations.CuboxAPILink)

		if err := client.SaveLink(entry.URL); err != nil {
			slog.Error("Unable to send entry to Cubox",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.ShioriEnabled {
		slog.Debug("Sending entry to Shiori",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := shiori.NewClient(
			userIntegrations.ShioriURL,
			userIntegrations.ShioriUsername,
			userIntegrations.ShioriPassword,
		)

		if err := client.CreateBookmark(entry.URL, entry.Title); err != nil {
			slog.Error("Unable to send entry to Shiori",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.ShaarliEnabled {
		slog.Debug("Sending entry to Shaarli",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := shaarli.NewClient(
			userIntegrations.ShaarliURL,
			userIntegrations.ShaarliAPISecret,
		)

		if err := client.CreateLink(entry.URL, entry.Title); err != nil {
			slog.Error("Unable to send entry to Shaarli",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.WebhookEnabled {
		slog.Debug("Sending entry to Webhook",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
			slog.String("webhook_url", userIntegrations.WebhookURL),
		)

		webhookClient := webhook.NewClient(userIntegrations.WebhookURL, userIntegrations.WebhookSecret)
		if err := webhookClient.SendSaveEntryWebhookEvent(entry); err != nil {
			slog.Error("Unable to send entry to Webhook",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.String("webhook_url", userIntegrations.WebhookURL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.OmnivoreEnabled {
		slog.Debug("Sending entry to Omnivore",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := omnivore.NewClient(userIntegrations.OmnivoreAPIKey, userIntegrations.OmnivoreURL)
		if err := client.SaveUrl(entry.URL); err != nil {
			slog.Error("Unable to send entry to Omnivore",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.KarakeepEnabled {
		slog.Debug("Sending entry to Karakeep",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := karakeep.NewClient(userIntegrations.KarakeepAPIKey, userIntegrations.KarakeepURL)
		if err := client.SaveURL(entry.URL); err != nil {
			slog.Error("Unable to send entry to Karakeep",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.RaindropEnabled {
		slog.Debug("Sending entry to Raindrop",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
		)

		client := raindrop.NewClient(userIntegrations.RaindropToken, userIntegrations.RaindropCollectionID, userIntegrations.RaindropTags)
		if err := client.CreateRaindrop(entry.URL, entry.Title); err != nil {
			slog.Error("Unable to send entry to Raindrop",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
		}
	}
}

// PushEntries pushes a list of entries to activated third-party providers during feed refreshes.
func PushEntries(feed *model.Feed, entries model.Entries, userIntegrations *model.Integration) {
	if userIntegrations.MatrixBotEnabled {
		slog.Debug("Sending new entries to Matrix",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		err := matrixbot.PushEntries(
			feed,
			entries,
			userIntegrations.MatrixBotURL,
			userIntegrations.MatrixBotUser,
			userIntegrations.MatrixBotPassword,
			userIntegrations.MatrixBotChatID,
		)
		if err != nil {
			slog.Error("Unable to send new entries to Matrix",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int("nb_entries", len(entries)),
				slog.Int64("feed_id", feed.ID),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.WebhookEnabled {
		var webhookURL string
		if feed.WebhookURL != "" {
			webhookURL = feed.WebhookURL
		} else {
			webhookURL = userIntegrations.WebhookURL
		}

		slog.Debug("Sending new entries to Webhook",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
			slog.String("webhook_url", webhookURL),
		)

		webhookClient := webhook.NewClient(webhookURL, userIntegrations.WebhookSecret)
		if err := webhookClient.SendNewEntriesWebhookEvent(feed, entries); err != nil {
			slog.Debug("Unable to send new entries to Webhook",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int("nb_entries", len(entries)),
				slog.Int64("feed_id", feed.ID),
				slog.String("webhook_url", webhookURL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.NtfyEnabled && feed.NtfyEnabled {
		ntfyTopic := feed.NtfyTopic
		if ntfyTopic == "" {
			ntfyTopic = userIntegrations.NtfyTopic
		}
		slog.Debug("Sending new entries to Ntfy",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
			slog.String("topic", ntfyTopic),
		)

		client := ntfy.NewClient(
			userIntegrations.NtfyURL,
			ntfyTopic,
			userIntegrations.NtfyAPIToken,
			userIntegrations.NtfyUsername,
			userIntegrations.NtfyPassword,
			userIntegrations.NtfyIconURL,
			userIntegrations.NtfyInternalLinks,
			feed.NtfyPriority,
		)

		if err := client.SendMessages(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Ntfy", slog.Any("error", err))
		}
	}

	if userIntegrations.AppriseEnabled {
		slog.Debug("Sending new entries to Apprise",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		appriseServiceURLs := userIntegrations.AppriseServicesURL
		if feed.AppriseServiceURLs != "" {
			appriseServiceURLs = feed.AppriseServiceURLs
		}

		client := apprise.NewClient(
			appriseServiceURLs,
			userIntegrations.AppriseURL,
		)

		if err := client.SendNotification(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Apprise", slog.Any("error", err))
		}
	}

	if userIntegrations.DiscordEnabled {
		slog.Debug("Sending new entries to Discord",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		client := discord.NewClient(
			userIntegrations.DiscordWebhookLink,
		)

		if err := client.SendDiscordMsg(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Discord", slog.Any("error", err))
		}
	}

	if userIntegrations.SlackEnabled {
		slog.Debug("Sending new entries to Slack",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		client := slack.NewClient(
			userIntegrations.SlackWebhookLink,
		)

		if err := client.SendSlackMsg(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Slack", slog.Any("error", err))
		}
	}

	if userIntegrations.PushoverEnabled && feed.PushoverEnabled {
		slog.Debug("Sending new entries to Pushover",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		client := pushover.New(
			userIntegrations.PushoverUser,
			userIntegrations.PushoverToken,
			feed.PushoverPriority,
			userIntegrations.PushoverDevice,
			userIntegrations.PushoverPrefix,
		)

		if err := client.SendMessages(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Pushover", slog.Any("error", err))
		}
	}

	// Integrations that only support sending individual entries
	if userIntegrations.TelegramBotEnabled {
		for _, entry := range entries {
			if userIntegrations.TelegramBotEnabled {
				slog.Debug("Sending a new entry to Telegram",
					slog.Int64("user_id", userIntegrations.UserID),
					slog.Int64("entry_id", entry.ID),
					slog.String("entry_url", entry.URL),
				)

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
					slog.Error("Unable to send entry to Telegram",
						slog.Int64("user_id", userIntegrations.UserID),
						slog.Int64("entry_id", entry.ID),
						slog.String("entry_url", entry.URL),
						slog.Any("error", err),
					)
				}
			}
		}
	}
}
