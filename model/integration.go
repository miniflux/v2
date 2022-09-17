// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

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
}
