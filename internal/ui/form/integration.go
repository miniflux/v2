// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
	"strconv"

	"miniflux.app/v2/internal/model"
)

// IntegrationForm represents user integration settings form.
type IntegrationForm struct {
	PinboardEnabled                  bool
	PinboardToken                    string
	PinboardTags                     string
	PinboardMarkAsUnread             bool
	InstapaperEnabled                bool
	InstapaperUsername               string
	InstapaperPassword               string
	FeverEnabled                     bool
	FeverUsername                    string
	FeverPassword                    string
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
	NotionEnabled                    bool
	NotionPageID                     string
	NotionToken                      string
	NunuxKeeperEnabled               bool
	NunuxKeeperURL                   string
	NunuxKeeperAPIKey                string
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
	LinkAceEnabled                   bool
	LinkAceURL                       string
	LinkAceAPIKey                    string
	LinkAceTags                      string
	LinkAcePrivate                   bool
	LinkAceCheckDisabled             bool
	LinkdingEnabled                  bool
	LinkdingURL                      string
	LinkdingAPIKey                   string
	LinkdingTags                     string
	LinkdingMarkAsUnread             bool
	LinkwardenEnabled                bool
	LinkwardenURL                    string
	LinkwardenAPIKey                 string
	MatrixBotEnabled                 bool
	MatrixBotUser                    string
	MatrixBotPassword                string
	MatrixBotURL                     string
	MatrixBotChatID                  string
	AppriseEnabled                   bool
	AppriseURL                       string
	AppriseServicesURL               string
	ReadeckEnabled                   bool
	ReadeckURL                       string
	ReadeckAPIKey                    string
	ReadeckLabels                    string
	ReadeckOnlyURL                   bool
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
	RSSBridgeEnabled                 bool
	RSSBridgeURL                     string
	OmnivoreEnabled                  bool
	OmnivoreAPIKey                   string
	OmnivoreURL                      string
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
	integration.WallabagOnlyURL = i.WallabagOnlyURL
	integration.WallabagURL = i.WallabagURL
	integration.WallabagClientID = i.WallabagClientID
	integration.WallabagClientSecret = i.WallabagClientSecret
	integration.WallabagUsername = i.WallabagUsername
	integration.WallabagPassword = i.WallabagPassword
	integration.NotionEnabled = i.NotionEnabled
	integration.NotionPageID = i.NotionPageID
	integration.NotionToken = i.NotionToken
	integration.NunuxKeeperEnabled = i.NunuxKeeperEnabled
	integration.NunuxKeeperURL = i.NunuxKeeperURL
	integration.NunuxKeeperAPIKey = i.NunuxKeeperAPIKey
	integration.EspialEnabled = i.EspialEnabled
	integration.EspialURL = i.EspialURL
	integration.EspialAPIKey = i.EspialAPIKey
	integration.EspialTags = i.EspialTags
	integration.ReadwiseEnabled = i.ReadwiseEnabled
	integration.ReadwiseAPIKey = i.ReadwiseAPIKey
	integration.PocketEnabled = i.PocketEnabled
	integration.PocketAccessToken = i.PocketAccessToken
	integration.PocketConsumerKey = i.PocketConsumerKey
	integration.TelegramBotEnabled = i.TelegramBotEnabled
	integration.TelegramBotToken = i.TelegramBotToken
	integration.TelegramBotChatID = i.TelegramBotChatID
	integration.TelegramBotTopicID = i.TelegramBotTopicID
	integration.TelegramBotDisableWebPagePreview = i.TelegramBotDisableWebPagePreview
	integration.TelegramBotDisableNotification = i.TelegramBotDisableNotification
	integration.TelegramBotDisableButtons = i.TelegramBotDisableButtons
	integration.LinkAceEnabled = i.LinkAceEnabled
	integration.LinkAceURL = i.LinkAceURL
	integration.LinkAceAPIKey = i.LinkAceAPIKey
	integration.LinkAceTags = i.LinkAceTags
	integration.LinkAcePrivate = i.LinkAcePrivate
	integration.LinkAceCheckDisabled = i.LinkAceCheckDisabled
	integration.LinkdingEnabled = i.LinkdingEnabled
	integration.LinkdingURL = i.LinkdingURL
	integration.LinkdingAPIKey = i.LinkdingAPIKey
	integration.LinkdingTags = i.LinkdingTags
	integration.LinkdingMarkAsUnread = i.LinkdingMarkAsUnread
	integration.LinkwardenEnabled = i.LinkwardenEnabled
	integration.LinkwardenURL = i.LinkwardenURL
	integration.LinkwardenAPIKey = i.LinkwardenAPIKey
	integration.MatrixBotEnabled = i.MatrixBotEnabled
	integration.MatrixBotUser = i.MatrixBotUser
	integration.MatrixBotPassword = i.MatrixBotPassword
	integration.MatrixBotURL = i.MatrixBotURL
	integration.MatrixBotChatID = i.MatrixBotChatID
	integration.AppriseEnabled = i.AppriseEnabled
	integration.AppriseServicesURL = i.AppriseServicesURL
	integration.AppriseURL = i.AppriseURL
	integration.ReadeckEnabled = i.ReadeckEnabled
	integration.ReadeckURL = i.ReadeckURL
	integration.ReadeckAPIKey = i.ReadeckAPIKey
	integration.ReadeckLabels = i.ReadeckLabels
	integration.ReadeckOnlyURL = i.ReadeckOnlyURL
	integration.ShioriEnabled = i.ShioriEnabled
	integration.ShioriURL = i.ShioriURL
	integration.ShioriUsername = i.ShioriUsername
	integration.ShioriPassword = i.ShioriPassword
	integration.ShaarliEnabled = i.ShaarliEnabled
	integration.ShaarliURL = i.ShaarliURL
	integration.ShaarliAPISecret = i.ShaarliAPISecret
	integration.WebhookEnabled = i.WebhookEnabled
	integration.WebhookURL = i.WebhookURL
	integration.RSSBridgeEnabled = i.RSSBridgeEnabled
	integration.RSSBridgeURL = i.RSSBridgeURL
	integration.OmnivoreEnabled = i.OmnivoreEnabled
	integration.OmnivoreAPIKey = i.OmnivoreAPIKey
	integration.OmnivoreURL = i.OmnivoreURL
}

// NewIntegrationForm returns a new IntegrationForm.
func NewIntegrationForm(r *http.Request) *IntegrationForm {
	return &IntegrationForm{
		PinboardEnabled:                  r.FormValue("pinboard_enabled") == "1",
		PinboardToken:                    r.FormValue("pinboard_token"),
		PinboardTags:                     r.FormValue("pinboard_tags"),
		PinboardMarkAsUnread:             r.FormValue("pinboard_mark_as_unread") == "1",
		InstapaperEnabled:                r.FormValue("instapaper_enabled") == "1",
		InstapaperUsername:               r.FormValue("instapaper_username"),
		InstapaperPassword:               r.FormValue("instapaper_password"),
		FeverEnabled:                     r.FormValue("fever_enabled") == "1",
		FeverUsername:                    r.FormValue("fever_username"),
		FeverPassword:                    r.FormValue("fever_password"),
		GoogleReaderEnabled:              r.FormValue("googlereader_enabled") == "1",
		GoogleReaderUsername:             r.FormValue("googlereader_username"),
		GoogleReaderPassword:             r.FormValue("googlereader_password"),
		WallabagEnabled:                  r.FormValue("wallabag_enabled") == "1",
		WallabagOnlyURL:                  r.FormValue("wallabag_only_url") == "1",
		WallabagURL:                      r.FormValue("wallabag_url"),
		WallabagClientID:                 r.FormValue("wallabag_client_id"),
		WallabagClientSecret:             r.FormValue("wallabag_client_secret"),
		WallabagUsername:                 r.FormValue("wallabag_username"),
		WallabagPassword:                 r.FormValue("wallabag_password"),
		NotionEnabled:                    r.FormValue("notion_enabled") == "1",
		NotionPageID:                     r.FormValue("notion_page_id"),
		NotionToken:                      r.FormValue("notion_token"),
		NunuxKeeperEnabled:               r.FormValue("nunux_keeper_enabled") == "1",
		NunuxKeeperURL:                   r.FormValue("nunux_keeper_url"),
		NunuxKeeperAPIKey:                r.FormValue("nunux_keeper_api_key"),
		EspialEnabled:                    r.FormValue("espial_enabled") == "1",
		EspialURL:                        r.FormValue("espial_url"),
		EspialAPIKey:                     r.FormValue("espial_api_key"),
		EspialTags:                       r.FormValue("espial_tags"),
		ReadwiseEnabled:                  r.FormValue("readwise_enabled") == "1",
		ReadwiseAPIKey:                   r.FormValue("readwise_api_key"),
		PocketEnabled:                    r.FormValue("pocket_enabled") == "1",
		PocketAccessToken:                r.FormValue("pocket_access_token"),
		PocketConsumerKey:                r.FormValue("pocket_consumer_key"),
		TelegramBotEnabled:               r.FormValue("telegram_bot_enabled") == "1",
		TelegramBotToken:                 r.FormValue("telegram_bot_token"),
		TelegramBotChatID:                r.FormValue("telegram_bot_chat_id"),
		TelegramBotTopicID:               optionalInt64Field(r.FormValue("telegram_bot_topic_id")),
		TelegramBotDisableWebPagePreview: r.FormValue("telegram_bot_disable_web_page_preview") == "1",
		TelegramBotDisableNotification:   r.FormValue("telegram_bot_disable_notification") == "1",
		TelegramBotDisableButtons:        r.FormValue("telegram_bot_disable_buttons") == "1",
		LinkAceEnabled:                   r.FormValue("linkace_enabled") == "1",
		LinkAceURL:                       r.FormValue("linkace_url"),
		LinkAceAPIKey:                    r.FormValue("linkace_api_key"),
		LinkAceTags:                      r.FormValue("linkace_tags"),
		LinkAcePrivate:                   r.FormValue("linkace_is_private") == "1",
		LinkAceCheckDisabled:             r.FormValue("linkace_check_disabled") == "1",
		LinkdingEnabled:                  r.FormValue("linkding_enabled") == "1",
		LinkdingURL:                      r.FormValue("linkding_url"),
		LinkdingAPIKey:                   r.FormValue("linkding_api_key"),
		LinkdingTags:                     r.FormValue("linkding_tags"),
		LinkdingMarkAsUnread:             r.FormValue("linkding_mark_as_unread") == "1",
		LinkwardenEnabled:                r.FormValue("linkwarden_enabled") == "1",
		LinkwardenURL:                    r.FormValue("linkwarden_url"),
		LinkwardenAPIKey:                 r.FormValue("linkwarden_api_key"),
		MatrixBotEnabled:                 r.FormValue("matrix_bot_enabled") == "1",
		MatrixBotUser:                    r.FormValue("matrix_bot_user"),
		MatrixBotPassword:                r.FormValue("matrix_bot_password"),
		MatrixBotURL:                     r.FormValue("matrix_bot_url"),
		MatrixBotChatID:                  r.FormValue("matrix_bot_chat_id"),
		AppriseEnabled:                   r.FormValue("apprise_enabled") == "1",
		AppriseURL:                       r.FormValue("apprise_url"),
		AppriseServicesURL:               r.FormValue("apprise_services_url"),
		ReadeckEnabled:                   r.FormValue("readeck_enabled") == "1",
		ReadeckURL:                       r.FormValue("readeck_url"),
		ReadeckAPIKey:                    r.FormValue("readeck_api_key"),
		ReadeckLabels:                    r.FormValue("readeck_labels"),
		ReadeckOnlyURL:                   r.FormValue("readeck_only_url") == "1",
		ShioriEnabled:                    r.FormValue("shiori_enabled") == "1",
		ShioriURL:                        r.FormValue("shiori_url"),
		ShioriUsername:                   r.FormValue("shiori_username"),
		ShioriPassword:                   r.FormValue("shiori_password"),
		ShaarliEnabled:                   r.FormValue("shaarli_enabled") == "1",
		ShaarliURL:                       r.FormValue("shaarli_url"),
		ShaarliAPISecret:                 r.FormValue("shaarli_api_secret"),
		WebhookEnabled:                   r.FormValue("webhook_enabled") == "1",
		WebhookURL:                       r.FormValue("webhook_url"),
		RSSBridgeEnabled:                 r.FormValue("rssbridge_enabled") == "1",
		RSSBridgeURL:                     r.FormValue("rssbridge_url"),
		OmnivoreEnabled:                  r.FormValue("omnivore_enabled") == "1",
		OmnivoreAPIKey:                   r.FormValue("omnivore_api_key"),
		OmnivoreURL:                      r.FormValue("omnivore_url"),
	}
}

func optionalInt64Field(formValue string) *int64 {
	if formValue == "" {
		return nil
	}
	value, _ := strconv.ParseInt(formValue, 10, 64)
	return &value
}
