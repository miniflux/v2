// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	timezones, err := h.store.Timezones()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	creds, err := h.store.WebAuthnCredentialsByUserID(loggedUser.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	settingsForm := form.NewSettingsForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", settingsForm)
	view.Set("readBehaviors", map[string]any{
		"NoAutoMarkAsRead":                           form.NoAutoMarkAsRead,
		"MarkAsReadOnView":                           form.MarkAsReadOnView,
		"MarkAsReadOnViewButWaitForPlayerCompletion": form.MarkAsReadOnViewButWaitForPlayerCompletion,
		"MarkAsReadOnlyOnPlayerCompletion":           form.MarkAsReadOnlyOnPlayerCompletion,
	})
	view.Set("themes", model.Themes())
	view.Set("languages", locale.AvailableLanguages)
	view.Set("timezones", timezones)
	view.Set("menu", "settings")
	view.Set("user", loggedUser)
	view.Set("countUnread", h.store.CountUnreadEntries(loggedUser.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID))
	view.Set("default_home_pages", model.HomePages())
	view.Set("categories_sorting_options", model.CategoriesSortingOptions())
	view.Set("countWebAuthnCerts", h.store.CountWebAuthnCredentialsByUserID(loggedUser.ID))
	view.Set("webAuthnCerts", creds)

	if validationErr := settingsForm.Validate(); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(loggedUser.Language))
		html.OK(w, r, view.Render("settings"))
		return
	}

	userModificationRequest := &model.UserModificationRequest{
		Username:               model.OptionalString(settingsForm.Username),
		Password:               model.OptionalString(settingsForm.Password),
		Theme:                  model.OptionalString(settingsForm.Theme),
		Language:               model.OptionalString(settingsForm.Language),
		Timezone:               model.OptionalString(settingsForm.Timezone),
		EntryDirection:         model.OptionalString(settingsForm.EntryDirection),
		EntryOrder:             model.OptionalString(settingsForm.EntryOrder),
		EntriesPerPage:         model.OptionalNumber(settingsForm.EntriesPerPage),
		CategoriesSortingOrder: model.OptionalString(settingsForm.CategoriesSortingOrder),
		DisplayMode:            model.OptionalString(settingsForm.DisplayMode),
		GestureNav:             model.OptionalString(settingsForm.GestureNav),
		DefaultReadingSpeed:    model.OptionalNumber(settingsForm.DefaultReadingSpeed),
		CJKReadingSpeed:        model.OptionalNumber(settingsForm.CJKReadingSpeed),
		DefaultHomePage:        model.OptionalString(settingsForm.DefaultHomePage),
		MediaPlaybackRate:      model.OptionalNumber(settingsForm.MediaPlaybackRate),
		BlockFilterEntryRules:  model.OptionalString(settingsForm.BlockFilterEntryRules),
		KeepFilterEntryRules:   model.OptionalString(settingsForm.KeepFilterEntryRules),
		ExternalFontHosts:      model.OptionalString(settingsForm.ExternalFontHosts),
	}

	if validationErr := validator.ValidateUserModification(h.store, loggedUser.ID, userModificationRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(loggedUser.Language))
		html.OK(w, r, view.Render("settings"))
		return
	}

	err = h.store.UpdateUser(settingsForm.Merge(loggedUser))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess.SetLanguage(loggedUser.Language)
	sess.SetTheme(loggedUser.Theme)
	sess.NewFlashMessage(locale.NewPrinter(request.UserLanguage(r)).Printf("alert.prefs_saved"))
	html.Redirect(w, r, route.Path(h.router, "settings"))
}
