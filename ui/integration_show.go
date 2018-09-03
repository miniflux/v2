// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// ShowIntegrations renders the page with all external integrations.
func (c *Controller) ShowIntegrations(w http.ResponseWriter, r *http.Request) {
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	integration, err := c.store.Integration(user.ID)
	if err != nil {
		html.ServerError(w, err)
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
		FeverPassword:        integration.FeverPassword,
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
	}

	sess := session.New(c.store, request.SessionID(r))
	view := view.New(c.tpl, r, sess)
	view.Set("form", integrationForm)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))
	view.Set("hasPocketConsumerKeyConfigured", c.cfg.PocketConsumerKey("") != "")

	html.OK(w, r, view.Render("integrations"))
}
