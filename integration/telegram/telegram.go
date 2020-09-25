// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package telegram // import "miniflux.app/integration/telegram"

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"miniflux.app/logger"
	"miniflux.app/storage"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SendTelegramMsg sends feed to Telegram.
func SendTelegramMsg(store *storage.Storage, userID int64, feedID int64, telegramItemMsg []string) {
	if len(telegramItemMsg) > 0 {
		integration, err := store.Integration(userID)
		if err != nil {
			logger.Error("[Telegram] %v", err)
			return
		}
		if integration != nil && integration.TelegramEnabled && len(integration.TelegramToken) > 0 {
			feed, storeErr := store.FeedByID(userID, feedID)
			if storeErr != nil {
				logger.Error("[Telegram] %v", storeErr)
				return
			}
			bot, botErr := tgbotapi.NewBotAPIWithClient(integration.TelegramToken, &http.Client{Timeout: 15 * time.Second})
			if botErr != nil {
				logger.Error("[Telegram] %v", botErr)
				return
			}
			if bot != nil {
				text := fmt.Sprintf("*%v*\n", feed.Title) + strings.Join(telegramItemMsg, "\n")
				chatID, parseErr := strconv.ParseInt(integration.TelegramChatID, 10, 64)
				if parseErr != nil {
					logger.Error("[Telegram] %v", parseErr)
					return
				}
				message := tgbotapi.NewMessage(chatID, text)
				message.DisableWebPagePreview = true
				message.ParseMode = "markdown"
				_, err := bot.Send(message)
				if err != nil {
					logger.Error(`[Telegram] feed #%d Send msg error %v`, feedID, err)
				}
			}
		}
	}
}
