// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/ui/form"
)

func (h *handler) updateIntegration(w http.ResponseWriter, r *http.Request) {
	sess := request.WebSession(r)
	printer := locale.NewPrinter(sess.Language())
	userID := request.UserID(r)

	integration, err := h.store.Integration(userID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	integrationForm := form.NewIntegrationForm(r)
	integrationForm.Merge(integration)

	if integration.FeverUsername != "" && h.store.HasDuplicateFeverUsername(userID, integration.FeverUsername) {
		sess.SetErrorMessage(printer.Print("error.duplicate_fever_username"))
		response.HTMLRedirect(w, r, h.routePath("/integrations"))
		return
	}

	if integration.FeverEnabled {
		if integrationForm.FeverPassword != "" {
			integration.FeverToken = fmt.Sprintf("%x", md5.Sum([]byte(integration.FeverUsername+":"+integrationForm.FeverPassword)))
		}
	} else {
		integration.FeverToken = ""
	}

	if integration.GoogleReaderUsername != "" && h.store.HasDuplicateGoogleReaderUsername(userID, integration.GoogleReaderUsername) {
		sess.SetErrorMessage(printer.Print("error.duplicate_googlereader_username"))
		response.HTMLRedirect(w, r, h.routePath("/integrations"))
		return
	}

	if integration.GoogleReaderEnabled {
		if integrationForm.GoogleReaderPassword != "" {
			integration.GoogleReaderPassword, err = crypto.HashPassword(integrationForm.GoogleReaderPassword)
			if err != nil {
				response.HTMLServerError(w, r, err)
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

	if integrationForm.LinktacoEnabled {
		if integrationForm.LinktacoAPIToken == "" || integrationForm.LinktacoOrgSlug == "" {
			sess.SetErrorMessage(printer.Print("error.linktaco_missing_required_fields"))
			response.HTMLRedirect(w, r, h.routePath("/integrations"))
			return
		}
		if integration.LinktacoVisibility == "" {
			integration.LinktacoVisibility = "PUBLIC"
		}
	}

	err = h.store.UpdateIntegration(integration)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	sess.SetSuccessMessage(printer.Print("alert.prefs_saved"))
	response.HTMLRedirect(w, r, h.routePath("/integrations"))
}
