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

	settingsForm := form.NewSettingsForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", settingsForm)
	view.Set("themes", model.Themes())
	view.Set("languages", locale.AvailableLanguages())
	view.Set("timezones", timezones)
	view.Set("menu", "settings")
	view.Set("user", loggedUser)
	view.Set("countUnread", h.store.CountUnreadEntries(loggedUser.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID))

	if validationErr := settingsForm.Validate(); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(loggedUser.Language))
		html.OK(w, r, view.Render("settings"))
		return
	}

	userModificationRequest := &model.UserModificationRequest{
		Username:            model.OptionalString(settingsForm.Username),
		Password:            model.OptionalString(settingsForm.Password),
		Theme:               model.OptionalString(settingsForm.Theme),
		Language:            model.OptionalString(settingsForm.Language),
		Timezone:            model.OptionalString(settingsForm.Timezone),
		EntryDirection:      model.OptionalString(settingsForm.EntryDirection),
		EntriesPerPage:      model.OptionalInt(settingsForm.EntriesPerPage),
		DisplayMode:         model.OptionalString(settingsForm.DisplayMode),
		GestureNav:          model.OptionalString(settingsForm.GestureNav),
		DefaultReadingSpeed: model.OptionalInt(settingsForm.DefaultReadingSpeed),
		CJKReadingSpeed:     model.OptionalInt(settingsForm.CJKReadingSpeed),
		DefaultHomePage:     model.OptionalString(settingsForm.DefaultHomePage),
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
