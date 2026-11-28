// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import "miniflux.app/v2/internal/crypto"

// googleReaderTokenSaltSize is the size in bytes of the random salt of the Google Reader token.
const googleReaderTokenSaltSize = 16

// Integration represents user integration settings.
type Integration struct {
	UserID                           int64
	BetulaEnabled                    bool
	BetulaURL                        string
	BetulaToken                      string
	PinboardEnabled                  bool
	PinboardToken                    string
	PinboardTags                     string
	PinboardMarkAsUnread             bool
	InstapaperEnabled                bool
	InstapaperUsername               string
	InstapaperPassword               string
	FeverEnabled                     bool
	FeverUsername                    string
	FeverTokenHash                   []byte
	GoogleReaderEnabled              bool
	GoogleReaderUsername             string
	GoogleReaderPassword             string // Legacy bcrypt hash, replaced by the token salt and hash at the next ClientLogin.
	GoogleReaderTokenSalt            []byte
	GoogleReaderTokenHash            []byte
	WallabagEnabled                  bool
	WallabagOnlyURL                  bool
	WallabagURL                      string
	WallabagClientID                 string
	WallabagClientSecret             string
	WallabagUsername                 string
	WallabagPassword                 string
	WallabagTags                     string
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
	LinktacoEnabled                  bool
	LinktacoAPIToken                 string
	LinktacoOrgSlug                  string
	LinktacoTags                     string
	LinktacoVisibility               string
	LinkwardenEnabled                bool
	LinkwardenURL                    string
	LinkwardenAPIKey                 string
	LinkwardenCollectionID           *int64
	MatrixBotEnabled                 bool
	MatrixBotUser                    string
	MatrixBotPassword                string
	MatrixBotURL                     string
	MatrixBotChatID                  string
	AppriseEnabled                   bool
	AppriseURL                       string
	AppriseServicesURL               string
	ReadeckEnabled                   bool
	ReadeckPushEnabled               bool
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
	RSSBridgeToken                   string
	OmnivoreEnabled                  bool
	OmnivoreAPIKey                   string
	OmnivoreURL                      string
	KarakeepEnabled                  bool
	KarakeepAPIKey                   string
	KarakeepURL                      string
	KarakeepTags                     string
	RaindropEnabled                  bool
	RaindropToken                    string
	RaindropCollectionID             string
	RaindropTags                     string
	NtfyEnabled                      bool
	NtfyTopic                        string
	NtfyURL                          string
	NtfyAPIToken                     string
	NtfyUsername                     string
	NtfyPassword                     string
	NtfyIconURL                      string
	NtfyInternalLinks                bool
	CuboxEnabled                     bool
	CuboxAPILink                     string
	DiscordEnabled                   bool
	DiscordWebhookLink               string
	SlackEnabled                     bool
	SlackWebhookLink                 string
	PushoverEnabled                  bool
	PushoverUser                     string
	PushoverToken                    string
	PushoverDevice                   string
	PushoverPrefix                   string
	ArchiveorgEnabled                bool
	GotifyEnabled                    bool
	GotifyURL                        string
	GotifyToken                      string
	GotifyPriority                   int
}

// SetGoogleReaderPassword derives a new Google Reader token secret from the password and keeps only its salt and SHA-256 digest.
func (i *Integration) SetGoogleReaderPassword(password string) (secret string) {
	i.GoogleReaderPassword = ""
	i.GoogleReaderTokenSalt = crypto.GenerateRandomBytes(googleReaderTokenSaltSize)
	secret = crypto.DerivePasswordToken(password, i.GoogleReaderTokenSalt)
	i.GoogleReaderTokenHash = crypto.HashToken(secret)
	return secret
}

// ClearGoogleReaderPassword removes the Google Reader credentials, which revokes the token.
func (i *Integration) ClearGoogleReaderPassword() {
	i.GoogleReaderPassword = ""
	i.GoogleReaderTokenSalt = nil
	i.GoogleReaderTokenHash = nil
}

// HasGoogleReaderPassword reports whether a Google Reader password is stored, in the current or the legacy form.
func (i *Integration) HasGoogleReaderPassword() bool {
	return len(i.GoogleReaderTokenHash) > 0 || i.GoogleReaderPassword != ""
}
