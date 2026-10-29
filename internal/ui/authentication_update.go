// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) updateAuthentication(w http.ResponseWriter, r *http.Request) {
	if config.Opts.DisableLocalAuth() {
		response.HTMLForbidden(w, r)
		return
	}

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	creds, err := h.store.WebAuthnCredentialsByUserID(user.ID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	authSettingsForm := form.NewAuthSettingsForm(r)

	view := view.New(h.tpl, r)
	view.Set("form", authSettingsForm)
	view.Set("menu", "settings")
	view.Set("user", user)
	navMetadata, _ := h.store.GetNavMetadata(user.ID)
	view.Set("countUnread", navMetadata.CountUnread)
	view.Set("countErrorFeeds", navMetadata.CountErrorFeeds)
	view.Set("countWebAuthnCerts", len(creds))
	view.Set("webAuthnCerts", creds)

	if validationErr := authSettingsForm.Validate(); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(user.Language))
		response.HTML(w, r, view.Render("authentication"))
		return
	}

	userModificationRequest := &model.UserModificationRequest{
		Username: model.OptionalString(authSettingsForm.Username),
		Password: model.OptionalString(authSettingsForm.Password),
	}

	if validationErr := validator.ValidateUserModification(h.store, user.ID, userModificationRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(user.Language))
		response.HTML(w, r, view.Render("authentication"))
		return
	}

	if err := h.store.UpdateUser(authSettingsForm.Merge(user)); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	sess := request.WebSession(r)

	if authSettingsForm.Password != "" {
		// Sign out every other session, and give this one a new identity in
		// case its cookie leaked along with the old password.
		if err := h.store.RemoveUserWebSessionsExcept(user.ID, sess.ID); err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		if err := authenticateWebSession(w, r, h.store, user); err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
	}

	sess.SetSuccessMessage(locale.NewPrinter(sess.Language()).Printf("alert.prefs_saved"))
	response.HTMLRedirect(w, r, h.routePath("/authentication"))
}
