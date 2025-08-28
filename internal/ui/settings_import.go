// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/timezone"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/validator"
	"miniflux.app/v2/internal/version"
)

func (h *handler) importSettings(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	defer r.Body.Close()
	inputFile, _, err := r.FormFile("import_file")
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	creds, err := h.store.WebAuthnCredentialsByUserID(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	settingsForm := form.SettingsForm{
		Username:                  user.Username,
		Theme:                     user.Theme,
		Language:                  user.Language,
		Timezone:                  user.Timezone,
		EntryDirection:            user.EntryDirection,
		EntryOrder:                user.EntryOrder,
		EntriesPerPage:            user.EntriesPerPage,
		KeyboardShortcuts:         user.KeyboardShortcuts,
		ShowReadingTime:           user.ShowReadingTime,
		CustomCSS:                 user.Stylesheet,
		CustomJS:                  user.CustomJS,
		ExternalFontHosts:         user.ExternalFontHosts,
		EntrySwipe:                user.EntrySwipe,
		GestureNav:                user.GestureNav,
		DisplayMode:               user.DisplayMode,
		DefaultReadingSpeed:       user.DefaultReadingSpeed,
		CJKReadingSpeed:           user.CJKReadingSpeed,
		DefaultHomePage:           user.DefaultHomePage,
		CategoriesSortingOrder:    user.CategoriesSortingOrder,
		MarkReadBehavior:          model.MarkAsReadBehavior(user.MarkReadOnView, user.MarkReadOnMediaPlayerCompletion),
		MediaPlaybackRate:         user.MediaPlaybackRate,
		BlockFilterEntryRules:     user.BlockFilterEntryRules,
		KeepFilterEntryRules:      user.KeepFilterEntryRules,
		AlwaysOpenExternalLinks:   user.AlwaysOpenExternalLinks,
		OpenExternalLinksInNewTab: user.OpenExternalLinksInNewTab,
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", settingsForm)
	view.Set("readBehaviors", map[string]any{
		"NoAutoMarkAsRead":                           model.NoAutoMarkAsRead,
		"MarkAsReadOnView":                           model.MarkAsReadOnView,
		"MarkAsReadOnViewButWaitForPlayerCompletion": model.MarkAsReadOnViewButWaitForPlayerCompletion,
		"MarkAsReadOnlyOnPlayerCompletion":           model.MarkAsReadOnlyOnPlayerCompletion,
	})
	view.Set("themes", model.Themes())
	view.Set("languages", locale.AvailableLanguages)
	view.Set("timezones", timezone.AvailableTimezones)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("default_home_pages", model.HomePages())
	view.Set("categories_sorting_options", model.CategoriesSortingOptions())
	view.Set("countWebAuthnCerts", h.store.CountWebAuthnCredentialsByUserID(user.ID))
	view.Set("webAuthnCerts", creds)

	userExport := &model.UserExport{}
	if err := json.NewDecoder(inputFile).Decode(userExport); err != nil {
		view.Set("errorMessage", err.Error()) // TODO: translate error message
		html.OK(w, r, view.Render("settings"))
		return
	}

	// The version field exists to allow for future changes to the user export format
	// but currently is not used.
	if userExport.Version != version.Version {
		slog.Warn("user settings import version mismatch", "imported version", userExport.Version)
	}
	userModificationRequest := &userExport.UserModificationRequest

	if validationErr := validator.ValidateUserModification(h.store, user.ID, userModificationRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(user.Language))
		html.OK(w, r, view.Render("settings"))
		return
	}

	// White-out certain fields that we never want updated through the importer
	userModificationRequest.Password = nil
	userModificationRequest.Username = nil

	userModificationRequest.Patch(user)
	err = h.store.UpdateUser(user)
	if err != nil {
		view.Set("errorMessage", err.Error()) // TODO: translate error message
		html.OK(w, r, view.Render("settings"))
		return
	}

	sess.SetLanguage(user.Language)
	sess.SetTheme(user.Theme)
	sess.NewFlashMessage(locale.NewPrinter(request.UserLanguage(r)).Printf("alert.prefs_saved"))
	html.Redirect(w, r, route.Path(h.router, "settings"))
}
