// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showSettingsPage(w http.ResponseWriter, r *http.Request) {
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
		GestureNav:             user.GestureNav,
		DisplayMode:            user.DisplayMode,
		DefaultReadingSpeed:    user.DefaultReadingSpeed,
		CJKReadingSpeed:        user.CJKReadingSpeed,
		DefaultHomePage:        user.DefaultHomePage,
		CategoriesSortingOrder: user.CategoriesSortingOrder,
		MarkReadOnView:         user.MarkReadOnView,
		MediaPlaybackRate:      user.MediaPlaybackRate,
	}

	timezones, err := h.store.Timezones()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	creds, err := h.store.WebAuthnCredentialsByUserID(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
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
	view.Set("countWebAuthnCerts", h.store.CountWebAuthnCredentialsByUserID(user.ID))
	view.Set("webAuthnCerts", creds)

	html.OK(w, r, view.Render("settings"))
}
