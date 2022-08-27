// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/locale"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
	"miniflux.app/validator"
)

func (h *handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

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

	view.Set("form", settingsForm)
	view.Set("themes", model.Themes())
	view.Set("languages", locale.AvailableLanguages())
	view.Set("timezones", timezones)
	view.Set("menu", "settings")
	view.Set("user", loggedUser)
	view.Set("countUnread", h.store.CountUnreadEntries(loggedUser.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID))

	if err := settingsForm.Validate(); err != nil {
		view.Set("errorMessage", err.Error())
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
		DefaultReadingSpeed: model.OptionalInt(settingsForm.DefaultReadingSpeed),
		CJKReadingSpeed:     model.OptionalInt(settingsForm.CJKReadingSpeed),
		DefaultHomePage:     model.OptionalString(settingsForm.DefaultHomePage),
	}

	if validationErr := validator.ValidateUserModification(h.store, loggedUser.ID, userModificationRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.TranslationKey)
		html.OK(w, r, view.Render("settings"))
		return
	}

	err = h.store.UpdateUser(settingsForm.Merge(loggedUser))
	if err != nil {
		logger.Error("[UI:UpdateSettings] %v", err)
		view.Set("errorMessage", "error.unable_to_update_user")
		html.OK(w, r, view.Render("settings"))
		return
	}

	sess.SetLanguage(loggedUser.Language)
	sess.SetTheme(loggedUser.Theme)
	sess.NewFlashMessage(locale.NewPrinter(request.UserLanguage(r)).Printf("alert.prefs_saved"))
	html.Redirect(w, r, route.Path(h.router, "settings"))
}
