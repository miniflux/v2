// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"miniflux.app/http/response"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
)

// UpdateIntegration updates integration settings.
func (c *Controller) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	sess := session.New(c.store, request.SessionID(r))
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

	integrationForm := form.NewIntegrationForm(r)
	integrationForm.Merge(integration)

	if integration.FeverUsername != "" && c.store.HasDuplicateFeverUsername(user.ID, integration.FeverUsername) {
		sess.NewFlashErrorMessage(c.translator.GetLanguage(request.UserLanguage(r)).Get("error.duplicate_fever_username"))
		response.Redirect(w, r, route.Path(c.router, "integrations"))
		return
	}

	if integration.FeverEnabled {
		integration.FeverToken = fmt.Sprintf("%x", md5.Sum([]byte(integration.FeverUsername+":"+integration.FeverPassword)))
	} else {
		integration.FeverToken = ""
	}

	err = c.store.UpdateIntegration(integration)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	sess.NewFlashMessage(c.translator.GetLanguage(request.UserLanguage(r)).Get("alert.prefs_saved"))
	response.Redirect(w, r, route.Path(c.router, "integrations"))
}
