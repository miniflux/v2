// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"

	"miniflux.app/model"
)

// IntegrationForm represents user integration settings form.
type IntegrationForm struct {
	PinboardEnabled      bool
	PinboardToken        string
	PinboardTags         string
	PinboardMarkAsUnread bool
	InstapaperEnabled    bool
	InstapaperUsername   string
	InstapaperPassword   string
	FeverEnabled         bool
	FeverUsername        string
	FeverPassword        string
	GoogleReaderEnabled  bool
	GoogleReaderUsername string
	GoogleReaderPassword string
	WallabagEnabled      bool
	WallabagURL          string
	WallabagClientID     string
	WallabagClientSecret string
	WallabagUsername     string
	WallabagPassword     string
	NunuxKeeperEnabled   bool
	NunuxKeeperURL       string
	NunuxKeeperAPIKey    string
	PocketEnabled        bool
	PocketAccessToken    string
	PocketConsumerKey    string
	TelegramBotEnabled   bool
	TelegramBotToken     string
	TelegramBotChatID    string
}

// Merge copy form values to the model.
func (i IntegrationForm) Merge(integration *model.Integration) {
	integration.PinboardEnabled = i.PinboardEnabled
	integration.PinboardToken = i.PinboardToken
	integration.PinboardTags = i.PinboardTags
	integration.PinboardMarkAsUnread = i.PinboardMarkAsUnread
	integration.InstapaperEnabled = i.InstapaperEnabled
	integration.InstapaperUsername = i.InstapaperUsername
	integration.InstapaperPassword = i.InstapaperPassword
	integration.FeverEnabled = i.FeverEnabled
	integration.FeverUsername = i.FeverUsername
	integration.GoogleReaderEnabled = i.GoogleReaderEnabled
	integration.GoogleReaderUsername = i.GoogleReaderUsername
	integration.WallabagEnabled = i.WallabagEnabled
	integration.WallabagURL = i.WallabagURL
	integration.WallabagClientID = i.WallabagClientID
	integration.WallabagClientSecret = i.WallabagClientSecret
	integration.WallabagUsername = i.WallabagUsername
	integration.WallabagPassword = i.WallabagPassword
	integration.NunuxKeeperEnabled = i.NunuxKeeperEnabled
	integration.NunuxKeeperURL = i.NunuxKeeperURL
	integration.NunuxKeeperAPIKey = i.NunuxKeeperAPIKey
	integration.PocketEnabled = i.PocketEnabled
	integration.PocketAccessToken = i.PocketAccessToken
	integration.PocketConsumerKey = i.PocketConsumerKey
	integration.TelegramBotEnabled = i.TelegramBotEnabled
	integration.TelegramBotToken = i.TelegramBotToken
	integration.TelegramBotChatID = i.TelegramBotChatID
}

// NewIntegrationForm returns a new IntegrationForm.
func NewIntegrationForm(r *http.Request) *IntegrationForm {
	return &IntegrationForm{
		PinboardEnabled:      r.FormValue("pinboard_enabled") == "1",
		PinboardToken:        r.FormValue("pinboard_token"),
		PinboardTags:         r.FormValue("pinboard_tags"),
		PinboardMarkAsUnread: r.FormValue("pinboard_mark_as_unread") == "1",
		InstapaperEnabled:    r.FormValue("instapaper_enabled") == "1",
		InstapaperUsername:   r.FormValue("instapaper_username"),
		InstapaperPassword:   r.FormValue("instapaper_password"),
		FeverEnabled:         r.FormValue("fever_enabled") == "1",
		FeverUsername:        r.FormValue("fever_username"),
		FeverPassword:        r.FormValue("fever_password"),
		GoogleReaderEnabled:  r.FormValue("googlereader_enabled") == "1",
		GoogleReaderUsername: r.FormValue("googlereader_username"),
		GoogleReaderPassword: r.FormValue("googlereader_password"),
		WallabagEnabled:      r.FormValue("wallabag_enabled") == "1",
		WallabagURL:          r.FormValue("wallabag_url"),
		WallabagClientID:     r.FormValue("wallabag_client_id"),
		WallabagClientSecret: r.FormValue("wallabag_client_secret"),
		WallabagUsername:     r.FormValue("wallabag_username"),
		WallabagPassword:     r.FormValue("wallabag_password"),
		NunuxKeeperEnabled:   r.FormValue("nunux_keeper_enabled") == "1",
		NunuxKeeperURL:       r.FormValue("nunux_keeper_url"),
		NunuxKeeperAPIKey:    r.FormValue("nunux_keeper_api_key"),
		PocketEnabled:        r.FormValue("pocket_enabled") == "1",
		PocketAccessToken:    r.FormValue("pocket_access_token"),
		PocketConsumerKey:    r.FormValue("pocket_consumer_key"),
		TelegramBotEnabled:   r.FormValue("telegram_bot_enabled") == "1",
		TelegramBotToken:     r.FormValue("telegram_bot_token"),
		TelegramBotChatID:    r.FormValue("telegram_bot_chat_id"),
	}
}
