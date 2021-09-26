// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/config"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
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
		PinboardEnabled:      integration.PinboardEnabled,
		PinboardToken:        integration.PinboardToken,
		PinboardTags:         integration.PinboardTags,
		PinboardMarkAsUnread: integration.PinboardMarkAsUnread,
		InstapaperEnabled:    integration.InstapaperEnabled,
		InstapaperUsername:   integration.InstapaperUsername,
		InstapaperPassword:   integration.InstapaperPassword,
		FeverEnabled:         integration.FeverEnabled,
		FeverUsername:        integration.FeverUsername,
		WallabagEnabled:      integration.WallabagEnabled,
		WallabagURL:          integration.WallabagURL,
		WallabagClientID:     integration.WallabagClientID,
		WallabagClientSecret: integration.WallabagClientSecret,
		WallabagUsername:     integration.WallabagUsername,
		WallabagPassword:     integration.WallabagPassword,
		NunuxKeeperEnabled:   integration.NunuxKeeperEnabled,
		NunuxKeeperURL:       integration.NunuxKeeperURL,
		NunuxKeeperAPIKey:    integration.NunuxKeeperAPIKey,
		PocketEnabled:        integration.PocketEnabled,
		PocketAccessToken:    integration.PocketAccessToken,
		PocketConsumerKey:    integration.PocketConsumerKey,
		TelegramBotEnabled:   integration.TelegramBotEnabled,
		TelegramBotToken:     integration.TelegramBotToken,
		TelegramBotChatID:    integration.TelegramBotChatID,
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
