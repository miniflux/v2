// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package integration // import "miniflux.app/integration"

import (
	"miniflux.app/config"
	"miniflux.app/integration/instapaper"
	"miniflux.app/integration/nunuxkeeper"
	"miniflux.app/integration/pinboard"
	"miniflux.app/integration/pocket"
	"miniflux.app/integration/telegrambot"
	"miniflux.app/integration/wallabag"
	"miniflux.app/logger"
	"miniflux.app/model"
)

// SendEntry sends the entry to third-party providers when the user click on "Save".
func SendEntry(entry *model.Entry, integration *model.Integration) {
	if integration.PinboardEnabled {
		logger.Debug("[Integration] Sending Entry #%d %q for User #%d to Pinboard", entry.ID, entry.URL, integration.UserID)

		client := pinboard.NewClient(integration.PinboardToken)
		err := client.AddBookmark(
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
		logger.Debug("[Integration] Sending Entry #%d %q for User #%d to Instapaper", entry.ID, entry.URL, integration.UserID)

		client := instapaper.NewClient(integration.InstapaperUsername, integration.InstapaperPassword)
		if err := client.AddURL(entry.URL, entry.Title); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.WallabagEnabled {
		logger.Debug("[Integration] Sending Entry #%d %q for User #%d to Wallabag", entry.ID, entry.URL, integration.UserID)

		client := wallabag.NewClient(
			integration.WallabagURL,
			integration.WallabagClientID,
			integration.WallabagClientSecret,
			integration.WallabagUsername,
			integration.WallabagPassword,
		)

		if err := client.AddEntry(entry.URL, entry.Title, entry.Content); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.NunuxKeeperEnabled {
		logger.Debug("[Integration] Sending Entry #%d %q for User #%d to NunuxKeeper", entry.ID, entry.URL, integration.UserID)

		client := nunuxkeeper.NewClient(
			integration.NunuxKeeperURL,
			integration.NunuxKeeperAPIKey,
		)

		if err := client.AddEntry(entry.URL, entry.Title, entry.Content); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}

	if integration.PocketEnabled {
		logger.Debug("[Integration] Sending Entry #%d %q for User #%d to Pocket", entry.ID, entry.URL, integration.UserID)

		client := pocket.NewClient(config.Opts.PocketConsumerKey(integration.PocketConsumerKey), integration.PocketAccessToken)
		if err := client.AddURL(entry.URL, entry.Title); err != nil {
			logger.Error("[Integration] UserID #%d: %v", integration.UserID, err)
		}
	}
}

// PushEntry pushes an entry to third-party providers during feed refreshes.
func PushEntry(entry *model.Entry, integration *model.Integration) {
	if integration.TelegramBotEnabled {
		logger.Debug("[Integration] Sending Entry %q for User #%d to Telegram", entry.URL, integration.UserID)

		err := telegrambot.PushEntry(entry, integration.TelegramBotToken, integration.TelegramBotChatID)
		if err != nil {
			logger.Error("[Integration] push entry to telegram bot failed: %v", err)
		}
	}
}
