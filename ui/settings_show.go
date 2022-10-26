// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/locale"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) showSettingsPage(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	settingsForm := form.SettingsForm{
		Username:               user.Username,
		Theme:                  user.Theme,
		Language:               user.Language,
		Timezone:               user.Timezone,
		EntryDirection:         user.EntryDirection,
		EntryOrder:             user.EntryOrder,
		EntriesPerPage:         user.EntriesPerPage,
		KeyboardShortcuts:      user.KeyboardShortcuts,
		ShowReadingTime:        user.ShowReadingTime,
		CustomCSS:              user.Stylesheet,
		EntrySwipe:             user.EntrySwipe,
		DisplayMode:            user.DisplayMode,
		DefaultReadingSpeed:    user.DefaultReadingSpeed,
		CJKReadingSpeed:        user.CJKReadingSpeed,
		DefaultHomePage:        user.DefaultHomePage,
		CategoriesSortingOrder: user.CategoriesSortingOrder,
	}

	timezones, err := h.store.Timezones()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	view.Set("form", settingsForm)
	view.Set("themes", model.Themes())
	view.Set("languages", locale.AvailableLanguages())
	view.Set("timezones", timezones)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("default_home_pages", model.HomePages())
	view.Set("categories_sorting_options", model.CategoriesSortingOptions())

	html.OK(w, r, view.Render("settings"))
}
