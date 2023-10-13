// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

// Integration represents user integration settings.
type Integration struct {
	UserID                           int64
	PinboardEnabled                  bool
	PinboardToken                    string
	PinboardTags                     string
	PinboardMarkAsUnread             bool
	InstapaperEnabled                bool
	InstapaperUsername               string
	InstapaperPassword               string
	FeverEnabled                     bool
	FeverUsername                    string
	FeverToken                       string
	GoogleReaderEnabled              bool
	GoogleReaderUsername             string
	GoogleReaderPassword             string
	WallabagEnabled                  bool
	WallabagOnlyURL                  bool
	WallabagURL                      string
	WallabagClientID                 string
	WallabagClientSecret             string
	WallabagUsername                 string
	WallabagPassword                 string
	NunuxKeeperEnabled               bool
	NunuxKeeperURL                   string
	NunuxKeeperAPIKey                string
	NotionEnabled                    bool
	NotionToken                      string
	NotionPageID                     string
	EspialEnabled                    bool
	EspialURL                        string
	EspialAPIKey                     string
	EspialTags                       string
	ReadwiseEnabled                  bool
	ReadwiseAPIKey                   string
	PocketEnabled                    bool
	PocketAccessToken                string
	PocketConsumerKey                string
	TelegramBotEnabled               bool
	TelegramBotToken                 string
	TelegramBotChatID                string
	TelegramBotTopicID               *int64
	TelegramBotDisableWebPagePreview bool
	TelegramBotDisableNotification   bool
	TelegramBotDisableButtons        bool
	LinkdingEnabled                  bool
	LinkdingURL                      string
	LinkdingAPIKey                   string
	LinkdingTags                     string
	LinkdingMarkAsUnread             bool
	MatrixBotEnabled                 bool
	MatrixBotUser                    string
	MatrixBotPassword                string
	MatrixBotURL                     string
	MatrixBotChatID                  string
	AppriseEnabled                   bool
	AppriseURL                       string
	AppriseServicesURL               string
	ShioriEnabled                    bool
	ShioriURL                        string
	ShioriUsername                   string
	ShioriPassword                   string
	ShaarliEnabled                   bool
	ShaarliURL                       string
	ShaarliAPISecret                 string
	WebhookEnabled                   bool
	WebhookURL                       string
	WebhookSecret                    string
}
