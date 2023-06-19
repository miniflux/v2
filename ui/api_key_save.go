// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) saveAPIKey(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	apiKeyForm := form.NewAPIKeyForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", apiKeyForm)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	if err := apiKeyForm.Validate(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("create_api_key"))
		return
	}

	if h.store.APIKeyExists(user.ID, apiKeyForm.Description) {
		view.Set("errorMessage", "error.api_key_already_exists")
		html.OK(w, r, view.Render("create_api_key"))
		return
	}

	apiKey := model.NewAPIKey(user.ID, apiKeyForm.Description)
	if err = h.store.CreateAPIKey(apiKey); err != nil {
		logger.Error("[UI:SaveAPIKey] %v", err)
		view.Set("errorMessage", "error.unable_to_create_api_key")
		html.OK(w, r, view.Render("create_api_key"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "apiKeys"))
}
