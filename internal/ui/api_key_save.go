// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) saveAPIKey(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	apiKeyForm := form.NewAPIKeyForm(r)
	apiKeyCreationRequest := &model.APIKeyCreationRequest{
		Description: apiKeyForm.Description,
	}

	if validationErr := validator.ValidateAPIKeyCreation(h.store, loggedUser.ID, apiKeyCreationRequest); validationErr != nil {
		sess := session.New(h.store, request.SessionID(r))
		view := view.New(h.tpl, r, sess)
		view.Set("form", apiKeyForm)
		view.Set("menu", "settings")
		view.Set("user", loggedUser)
		view.Set("countUnread", h.store.CountUnreadEntries(loggedUser.ID))
		view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID))
		view.Set("errorMessage", validationErr.Translate(loggedUser.Language))
		html.OK(w, r, view.Render("create_api_key"))
		return
	}

	if _, err = h.store.CreateAPIKey(loggedUser.ID, apiKeyCreationRequest.Description); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "apiKeys"))
}
