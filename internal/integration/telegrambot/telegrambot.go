// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telegrambot // import "miniflux.app/v2/internal/integration/telegrambot"

import (
	"fmt"
	"log/slog"
	"strconv"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"
)

func PushEntry(feed *model.Feed, entry *model.Entry, botToken, chatID string, topicID *int64, disableWebPagePreview, disableNotification bool, disableButtons bool) error {
	formattedText := fmt.Sprintf(
		`<b>%s</b> - <a href=%q>%s</a>`,
		feed.Title,
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

		baseURL := config.Opts.BaseURL()
		entryPath := "/unread/entry/" + strconv.FormatInt(entry.ID, 10)

		minifluxEntryURL, err := urllib.JoinBaseURLAndPath(baseURL, entryPath)
		if err != nil {
			slog.Error("Unable to create Miniflux entry URL", slog.Any("error", err))
		} else {
			minifluxEntryURLButton := InlineKeyboardButton{Text: "Go to Miniflux", URL: minifluxEntryURL}
			markupRow = append(markupRow, &minifluxEntryURLButton)
		}

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
