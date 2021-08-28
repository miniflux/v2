package telegrambot

import (
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"html/template"
	"miniflux.app/model"
	"strconv"
)

func PushEntry(entry *model.Entry, integration *model.Integration) error {
	if !integration.TelegramBotEnabled {
		return nil
	}

	bot, err := tgbotapi.NewBotAPI(integration.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("create bot failed: %w", err)
	}

	t, err := template.New("message").Parse("{{ .Title }}\n<a href=\"{{ .URL }}\">{{ .URL }}</a>")
	if err != nil {
		return fmt.Errorf("parse template failed: %w", err)
	}

	var result bytes.Buffer

	err = t.Execute(&result, entry)
	if err != nil {
		return fmt.Errorf("execute template failed: %w", err)
	}

	chatId, _ := strconv.ParseInt(integration.TelegramBotChatID, 10, 64)
	msg := tgbotapi.NewMessage(chatId, result.String())
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = false
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("send message failed: %w", err)
	}

	return nil
}
