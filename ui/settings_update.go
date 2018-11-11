// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/response/html"
	"miniflux.app/http/request"
	"miniflux.app/http/route"
	"miniflux.app/locale"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
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
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))

	if err := settingsForm.Validate(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("settings"))
		return
	}

	if h.store.AnotherUserExists(user.ID, settingsForm.Username) {
		view.Set("errorMessage", "error.user_already_exists")
		html.OK(w, r, view.Render("settings"))
		return
	}

	err = h.store.UpdateUser(settingsForm.Merge(user))
	if err != nil {
		logger.Error("[UI:UpdateSettings] %v", err)
		view.Set("errorMessage", "error.unable_to_update_user")
		html.OK(w, r, view.Render("settings"))
		return
	}

	sess.SetLanguage(user.Language)
	sess.SetTheme(user.Theme)
	sess.NewFlashMessage(locale.NewPrinter(request.UserLanguage(r)).Printf("alert.prefs_saved"))
	html.Redirect(w, r, route.Path(h.router, "settings"))
}
