package telegrambot

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
		return fmt.Errorf("telegrambot: create bot failed: %w", err)
	}

	t, err := template.New("message").Parse("{{ .Title }}\n<a href=\"{{ .URL }}\">{{ .URL }}</a>")
	if err != nil {
		return fmt.Errorf("telegrambot: parse template failed: %w", err)
	}

	var result bytes.Buffer

	err = t.Execute(&result, entry)
	if err != nil {
		return fmt.Errorf("telegrambot: execute template failed: %w", err)
	}

	chatId, _ := strconv.ParseInt(chatID, 10, 64)
	msg := tgbotapi.NewMessage(chatId, result.String())
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = false
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("telegrambot: send message failed: %w", err)
	}

	return nil
}
