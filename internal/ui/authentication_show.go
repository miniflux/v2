// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showAuthenticationPage(w http.ResponseWriter, r *http.Request) {
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

	hasPassword, err := h.store.HasPassword(user.ID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	view := view.New(h.tpl, r)
	view.Set("form", form.AuthSettingsForm{Username: user.Username})
	view.Set("hasPassword", hasPassword)
	view.Set("menu", "settings")
	view.Set("user", user)
	navMetadata, _ := h.store.GetNavMetadata(user.ID)
	view.Set("countUnread", navMetadata.CountUnread)
	view.Set("countErrorFeeds", navMetadata.CountErrorFeeds)
	view.Set("countWebAuthnCerts", len(creds))
	view.Set("webAuthnCerts", creds)

	response.HTML(w, r, view.Render("authentication"))
}
