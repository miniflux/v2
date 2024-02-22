// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showIntegrationPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	integration, err := h.store.Integration(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	integrationForm := form.IntegrationForm{
		PinboardEnabled:                  integration.PinboardEnabled,
		PinboardToken:                    integration.PinboardToken,
		PinboardTags:                     integration.PinboardTags,
		PinboardMarkAsUnread:             integration.PinboardMarkAsUnread,
		InstapaperEnabled:                integration.InstapaperEnabled,
		InstapaperUsername:               integration.InstapaperUsername,
		InstapaperPassword:               integration.InstapaperPassword,
		FeverEnabled:                     integration.FeverEnabled,
		FeverUsername:                    integration.FeverUsername,
		GoogleReaderEnabled:              integration.GoogleReaderEnabled,
		GoogleReaderUsername:             integration.GoogleReaderUsername,
		WallabagEnabled:                  integration.WallabagEnabled,
		WallabagOnlyURL:                  integration.WallabagOnlyURL,
		WallabagURL:                      integration.WallabagURL,
		WallabagClientID:                 integration.WallabagClientID,
		WallabagClientSecret:             integration.WallabagClientSecret,
		WallabagUsername:                 integration.WallabagUsername,
		WallabagPassword:                 integration.WallabagPassword,
		NotionEnabled:                    integration.NotionEnabled,
		NotionPageID:                     integration.NotionPageID,
		NotionToken:                      integration.NotionToken,
		NunuxKeeperEnabled:               integration.NunuxKeeperEnabled,
		NunuxKeeperURL:                   integration.NunuxKeeperURL,
		NunuxKeeperAPIKey:                integration.NunuxKeeperAPIKey,
		EspialEnabled:                    integration.EspialEnabled,
		EspialURL:                        integration.EspialURL,
		EspialAPIKey:                     integration.EspialAPIKey,
		EspialTags:                       integration.EspialTags,
		ReadwiseEnabled:                  integration.ReadwiseEnabled,
		ReadwiseAPIKey:                   integration.ReadwiseAPIKey,
		PocketEnabled:                    integration.PocketEnabled,
		PocketAccessToken:                integration.PocketAccessToken,
		PocketConsumerKey:                integration.PocketConsumerKey,
		TelegramBotEnabled:               integration.TelegramBotEnabled,
		TelegramBotToken:                 integration.TelegramBotToken,
		TelegramBotChatID:                integration.TelegramBotChatID,
		TelegramBotTopicID:               integration.TelegramBotTopicID,
		TelegramBotDisableWebPagePreview: integration.TelegramBotDisableWebPagePreview,
		TelegramBotDisableNotification:   integration.TelegramBotDisableNotification,
		TelegramBotDisableButtons:        integration.TelegramBotDisableButtons,
		LinkAceEnabled:                   integration.LinkAceEnabled,
		LinkAceURL:                       integration.LinkAceURL,
		LinkAceAPIKey:                    integration.LinkAceAPIKey,
		LinkAceTags:                      integration.LinkAceTags,
		LinkAcePrivate:                   integration.LinkAcePrivate,
		LinkAceCheckDisabled:             integration.LinkAceCheckDisabled,
		LinkdingEnabled:                  integration.LinkdingEnabled,
		LinkdingURL:                      integration.LinkdingURL,
		LinkdingAPIKey:                   integration.LinkdingAPIKey,
		LinkdingTags:                     integration.LinkdingTags,
		LinkdingMarkAsUnread:             integration.LinkdingMarkAsUnread,
		LinkwardenEnabled:                integration.LinkwardenEnabled,
		LinkwardenURL:                    integration.LinkwardenURL,
		LinkwardenAPIKey:                 integration.LinkwardenAPIKey,
		MatrixBotEnabled:                 integration.MatrixBotEnabled,
		MatrixBotUser:                    integration.MatrixBotUser,
		MatrixBotPassword:                integration.MatrixBotPassword,
		MatrixBotURL:                     integration.MatrixBotURL,
		MatrixBotChatID:                  integration.MatrixBotChatID,
		AppriseEnabled:                   integration.AppriseEnabled,
		AppriseURL:                       integration.AppriseURL,
		AppriseServicesURL:               integration.AppriseServicesURL,
		ReadeckEnabled:                   integration.ReadeckEnabled,
		ReadeckURL:                       integration.ReadeckURL,
		ReadeckAPIKey:                    integration.ReadeckAPIKey,
		ReadeckLabels:                    integration.ReadeckLabels,
		ReadeckOnlyURL:                   integration.ReadeckOnlyURL,
		ShioriEnabled:                    integration.ShioriEnabled,
		ShioriURL:                        integration.ShioriURL,
		ShioriUsername:                   integration.ShioriUsername,
		ShioriPassword:                   integration.ShioriPassword,
		ShaarliEnabled:                   integration.ShaarliEnabled,
		ShaarliURL:                       integration.ShaarliURL,
		ShaarliAPISecret:                 integration.ShaarliAPISecret,
		WebhookEnabled:                   integration.WebhookEnabled,
		WebhookURL:                       integration.WebhookURL,
		WebhookSecret:                    integration.WebhookSecret,
		RSSBridgeEnabled:                 integration.RSSBridgeEnabled,
		RSSBridgeURL:                     integration.RSSBridgeURL,
		OmnivoreEnabled:                  integration.OmnivoreEnabled,
		OmnivoreAPIKey:                   integration.OmnivoreAPIKey,
		OmnivoreURL:                      integration.OmnivoreURL,
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", integrationForm)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasPocketConsumerKeyConfigured", config.Opts.PocketConsumerKey("") != "")

	html.OK(w, r, view.Render("integrations"))
}
