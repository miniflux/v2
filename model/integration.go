// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/model"

// Integration represents user integration settings.
type Integration struct {
	UserID               int64
	PinboardEnabled      bool
	PinboardToken        string
	PinboardTags         string
	PinboardMarkAsUnread bool
	InstapaperEnabled    bool
	InstapaperUsername   string
	InstapaperPassword   string
	FeverEnabled         bool
	FeverUsername        string
	FeverToken           string
	GoogleReaderEnabled  bool
	GoogleReaderUsername string
	GoogleReaderPassword string
	WallabagEnabled      bool
	WallabagOnlyURL      bool
	WallabagURL          string
	WallabagClientID     string
	WallabagClientSecret string
	WallabagUsername     string
	WallabagPassword     string
	NunuxKeeperEnabled   bool
	NunuxKeeperURL       string
	NunuxKeeperAPIKey    string
	EspialEnabled        bool
	EspialURL            string
	EspialAPIKey         string
	EspialTags           string
	PocketEnabled        bool
	PocketAccessToken    string
	PocketConsumerKey    string
	TelegramBotEnabled   bool
	TelegramBotToken     string
	TelegramBotChatID    string
	LinkdingEnabled      bool
	LinkdingURL          string
	LinkdingAPIKey       string
	LinkdingTags         string
	LinkdingMarkAsUnread bool
	MatrixBotEnabled     bool
	MatrixBotUser        string
	MatrixBotPassword    string
	MatrixBotURL         string
	MatrixBotChatID      string
}
