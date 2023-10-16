// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telegrambot // import "miniflux.app/v2/internal/integration/telegrambot"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"miniflux.app/v2/internal/version"
)

const (
	defaultClientTimeout = 10 * time.Second
	telegramAPIEndpoint  = "https://api.telegram.org"

	MarkdownFormatting   = "Markdown"
	MarkdownV2Formatting = "MarkdownV2"
	HTMLFormatting       = "HTML"
)

type Client struct {
	botToken string
	chatID   string
}

func NewClient(botToken, chatID string) *Client {
	return &Client{
		botToken: botToken,
		chatID:   chatID,
	}
}

// Specs: https://core.telegram.org/bots/api#getme
func (c *Client) GetMe() (*User, error) {
	endpointURL, err := url.JoinPath(telegramAPIEndpoint, "/bot"+c.botToken, "/getMe")
	if err != nil {
		return nil, fmt.Errorf("telegram: unable to join base URL and path: %w", err)
	}

	request, err := http.NewRequest(http.MethodGet, endpointURL, nil)
	if err != nil {
		return nil, fmt.Errorf("telegram: unable to create request: %v", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("telegram: unable to send request: %v", err)
	}
	defer response.Body.Close()

	var userResponse UserResponse
	if err := json.NewDecoder(response.Body).Decode(&userResponse); err != nil {
		return nil, fmt.Errorf("telegram: unable to decode user response: %w", err)
	}

	if !userResponse.Ok {
		return nil, fmt.Errorf("telegram: unable to send message: %s (error code is %d)", userResponse.Description, userResponse.ErrorCode)
	}

	return &userResponse.Result, nil
}

// Specs: https://core.telegram.org/bots/api#sendmessage
func (c *Client) SendMessage(message *MessageRequest) (*Message, error) {
	endpointURL, err := url.JoinPath(telegramAPIEndpoint, "/bot"+c.botToken, "/sendMessage")
	if err != nil {
		return nil, fmt.Errorf("telegram: unable to join base URL and path: %w", err)
	}

	requestBody, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("telegram: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, endpointURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("telegram: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("telegram: unable to send request: %v", err)
	}
	defer response.Body.Close()

	var messageResponse MessageResponse
	if err := json.NewDecoder(response.Body).Decode(&messageResponse); err != nil {
		return nil, fmt.Errorf("telegram: unable to decode discovery response: %w", err)
	}

	if !messageResponse.Ok {
		return nil, fmt.Errorf("telegram: unable to send message: %s (error code is %d)", messageResponse.Description, messageResponse.ErrorCode)
	}

	return &messageResponse.Result, nil
}

type InlineKeyboard struct {
	InlineKeyboard []InlineKeyboardRow `json:"inline_keyboard"`
}

type InlineKeyboardRow []*InlineKeyboardButton

type InlineKeyboardButton struct {
	Text string `json:"text"`
	URL  string `json:"url,omitempty"`
}

type User struct {
	ID                      int64  `json:"id"`
	IsBot                   bool   `json:"is_bot"`
	FirstName               string `json:"first_name"`
	LastName                string `json:"last_name"`
	Username                string `json:"username"`
	LanguageCode            string `json:"language_code"`
	IsPremium               bool   `json:"is_premium"`
	CanJoinGroups           bool   `json:"can_join_groups"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries"`
}

type Chat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type Message struct {
	MessageID       int64 `json:"message_id"`
	From            User  `json:"from"`
	Chat            Chat  `json:"chat"`
	MessageThreadID int64 `json:"message_thread_id"`
	Date            int64 `json:"date"`
}

type BaseResponse struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type UserResponse struct {
	BaseResponse
	Result User `json:"result"`
}

type MessageRequest struct {
	ChatID                string          `json:"chat_id"`
	MessageThreadID       int64           `json:"message_thread_id,omitempty"`
	Text                  string          `json:"text"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool            `json:"disable_web_page_preview"`
	DisableNotification   bool            `json:"disable_notification"`
	ReplyMarkup           *InlineKeyboard `json:"reply_markup,omitempty"`
}

type MessageResponse struct {
	BaseResponse
	Result Message `json:"result"`
}
