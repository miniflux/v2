// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
)

func (h *handler) updateIntegration(w http.ResponseWriter, r *http.Request) {
	printer := locale.NewPrinter(request.UserLanguage(r))
	sess := session.New(h.store, request.SessionID(r))
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	integration, err := h.store.Integration(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	integrationForm := form.NewIntegrationForm(r)
	integrationForm.Merge(integration)

	if integration.FeverUsername != "" && h.store.HasDuplicateFeverUsername(user.ID, integration.FeverUsername) {
		sess.NewFlashErrorMessage(printer.Print("error.duplicate_fever_username"))
		html.Redirect(w, r, route.Path(h.router, "integrations"))
		return
	}

	if integration.FeverEnabled {
		if integrationForm.FeverPassword != "" {
			integration.FeverToken = fmt.Sprintf("%x", md5.Sum([]byte(integration.FeverUsername+":"+integrationForm.FeverPassword)))
		}
	} else {
		integration.FeverToken = ""
	}

	if integration.GoogleReaderUsername != "" && h.store.HasDuplicateGoogleReaderUsername(user.ID, integration.GoogleReaderUsername) {
		sess.NewFlashErrorMessage(printer.Print("error.duplicate_googlereader_username"))
		html.Redirect(w, r, route.Path(h.router, "integrations"))
		return
	}

	if integration.GoogleReaderEnabled {
		if integrationForm.GoogleReaderPassword != "" {
			integration.GoogleReaderPassword, err = crypto.HashPassword(integrationForm.GoogleReaderPassword)
			if err != nil {
				html.ServerError(w, r, err)
				return
			}
		}
	} else {
		integration.GoogleReaderPassword = ""
	}

	if integrationForm.WebhookEnabled {
		if integrationForm.WebhookURL == "" {
			integration.WebhookEnabled = false
			integration.WebhookSecret = ""
		} else if integration.WebhookSecret == "" {
			integration.WebhookSecret = crypto.GenerateRandomStringHex(32)
		}
	} else {
		integration.WebhookURL = ""
		integration.WebhookSecret = ""
	}

	err = h.store.UpdateIntegration(integration)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess.NewFlashMessage(printer.Print("alert.prefs_saved"))
	html.Redirect(w, r, route.Path(h.router, "integrations"))
}
