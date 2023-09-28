// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telegrambot // import "miniflux.app/v2/internal/integration/telegrambot"

import (
	"fmt"

	"miniflux.app/v2/internal/model"
)

func PushEntry(feed *model.Feed, entry *model.Entry, botToken, chatID string, topicID *int64, disableWebPagePreview, disableNotification bool, disableButtons bool) error {
	formattedText := fmt.Sprintf(
		`<a href=%q>%s</a>`,
		entry.URL,
		entry.Title,
	)

	message := &MessageRequest{
		ChatID:                chatID,
		Text:                  formattedText,
		ParseMode:             HTMLFormatting,
		DisableWebPagePreview: disableWebPagePreview,
		DisableNotification:   disableNotification,
	}

	if topicID != nil {
		message.MessageThreadID = *topicID
	}

	if !disableButtons {
		var markupRow []*InlineKeyboardButton

		websiteURLButton := InlineKeyboardButton{Text: "Go to website", URL: feed.SiteURL}
		markupRow = append(markupRow, &websiteURLButton)

		articleURLButton := InlineKeyboardButton{Text: "Go to article", URL: entry.URL}
		markupRow = append(markupRow, &articleURLButton)

		if entry.CommentsURL != "" {
			commentURLButton := InlineKeyboardButton{Text: "Comments", URL: entry.CommentsURL}
			markupRow = append(markupRow, &commentURLButton)
		}

		message.ReplyMarkup = &InlineKeyboard{}
		message.ReplyMarkup.InlineKeyboard = append(message.ReplyMarkup.InlineKeyboard, markupRow)
	}

	client := NewClient(botToken, chatID)
	_, err := client.SendMessage(message)
	return err
}
