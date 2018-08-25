// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/session"
)

// OAuth2Redirect redirects the user to the consent page to ask for permission.
func (c *Controller) OAuth2Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	sess := session.New(c.store, ctx)

	provider := request.Param(r, "provider", "")
	if provider == "" {
		logger.Error("[OAuth2] Invalid or missing provider: %s", provider)
		response.Redirect(w, r, route.Path(c.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(c.cfg).Provider(provider)
	if err != nil {
		logger.Error("[OAuth2] %v", err)
		response.Redirect(w, r, route.Path(c.router, "login"))
		return
	}

	response.Redirect(w, r, authProvider.GetRedirectURL(sess.NewOAuth2State()))
}
