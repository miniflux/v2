// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "influxeed-engine/v2/internal/ui"

import (
	"net/http"

	"influxeed-engine/v2/internal/http/request"
	"influxeed-engine/v2/internal/http/response/html"
	"influxeed-engine/v2/internal/http/route"
	"influxeed-engine/v2/internal/locale"
	"influxeed-engine/v2/internal/model"
	"influxeed-engine/v2/internal/ui/form"
	"influxeed-engine/v2/internal/ui/session"
	"influxeed-engine/v2/internal/ui/view"
	"influxeed-engine/v2/internal/validator"
)

func (h *handler) saveUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if !user.IsAdmin {
		html.Forbidden(w, r)
		return
	}

	userForm := form.NewUserForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("form", userForm)

	if validationErr := userForm.ValidateCreation(); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(user.Language))
		html.OK(w, r, view.Render("create_user"))
		return
	}

	if h.store.UserExists(userForm.Username) {
		view.Set("errorMessage", locale.NewLocalizedError("error.user_already_exists").Translate(user.Language))
		html.OK(w, r, view.Render("create_user"))
		return
	}

	userCreationRequest := &model.UserCreationRequest{
		Username: userForm.Username,
		Password: userForm.Password,
		IsAdmin:  userForm.IsAdmin,
	}

	if validationErr := validator.ValidateUserCreationWithPassword(h.store, userCreationRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(user.Language))
		html.OK(w, r, view.Render("create_user"))
		return
	}

	if _, err := h.store.CreateUser(userCreationRequest); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "users"))
}
