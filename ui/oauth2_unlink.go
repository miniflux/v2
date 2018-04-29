// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/http/route"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/ui/session"
)

// OAuth2Unlink unlink an account from the external provider.
func (c *Controller) OAuth2Unlink(w http.ResponseWriter, r *http.Request) {
	provider := request.Param(r, "provider", "")
	if provider == "" {
		logger.Info("[OAuth2] Invalid or missing provider")
		response.Redirect(w, r, route.Path(c.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(c.cfg).Provider(provider)
	if err != nil {
		logger.Error("[OAuth2] %v", err)
		response.Redirect(w, r, route.Path(c.router, "settings"))
		return
	}

	ctx := context.New(r)
	if err := c.store.RemoveExtraField(ctx.UserID(), authProvider.GetUserExtraKey()); err != nil {
		html.ServerError(w, err)
		return
	}

	sess := session.New(c.store, ctx)
	sess.NewFlashMessage(c.translator.GetLanguage(ctx.UserLanguage()).Get("Your external account is now dissociated!"))
	response.Redirect(w, r, route.Path(c.router, "settings"))
	return
}
