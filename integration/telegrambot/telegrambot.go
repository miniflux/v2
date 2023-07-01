// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telegrambot // import "miniflux.app/integration/telegrambot"

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"miniflux.app/model"
)

// PushEntry pushes entry to telegram chat using integration settings provided
func PushEntry(entry *model.Entry, botToken, chatID string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("telegrambot: bot creation failed: %w", err)
	}

	tpl, err := template.New("message").Parse("{{ .Title }}\n<a href=\"{{ .URL }}\">{{ .URL }}</a>")
	if err != nil {
		return fmt.Errorf("telegrambot: template parsing failed: %w", err)
	}

	var result bytes.Buffer
	if err := tpl.Execute(&result, entry); err != nil {
		return fmt.Errorf("telegrambot: template execution failed: %w", err)
	}

	chatIDInt, _ := strconv.ParseInt(chatID, 10, 64)
	msg := tgbotapi.NewMessage(chatIDInt, result.String())
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = false

	if entry.CommentsURL != "" {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("Comments", entry.CommentsURL),
			))
	}

	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("telegrambot: sending message failed: %w", err)
	}

	return nil
}
