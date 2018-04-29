// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/ui/form"
	"github.com/miniflux/miniflux/ui/session"
	"github.com/miniflux/miniflux/ui/view"
)

// ShowIntegrations renders the page with all external integrations.
func (c *Controller) ShowIntegrations(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)

	user, err := c.store.UserByID(ctx.UserID())
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
	}

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("form", integrationForm)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))

	html.OK(w, view.Render("integrations"))
}
