// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
	"strconv"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/validator"
)

// SettingsForm represents the settings form.
type SettingsForm struct {
	Username               string
	Password               string
	Confirmation           string
	Theme                  string
	Language               string
	Timezone               string
	EntryDirection         string
	EntryOrder             string
	EntriesPerPage         int
	KeyboardShortcuts      bool
	ShowReadingTime        bool
	CustomCSS              string
	CustomJS               string
	ExternalFontHosts      string
	EntrySwipe             bool
	GestureNav             string
	DisplayMode            string
	DefaultReadingSpeed    int
	CJKReadingSpeed        int
	DefaultHomePage        string
	CategoriesSortingOrder string
	MarkReadOnView         bool
	// MarkReadBehavior is a string representation of the MarkReadOnView and MarkReadOnMediaPlayerCompletion fields together
	MarkReadBehavior          model.MarkReadBehavior
	MediaPlaybackRate         float64
	BlockFilterEntryRules     string
	KeepFilterEntryRules      string
	AlwaysOpenExternalLinks   bool
	OpenExternalLinksInNewTab bool
}

// Validate makes sure the form values are valid.
func (s *SettingsForm) Validate() *locale.LocalizedError {
	if (s.Username == "" && !config.Opts.DisableLocalAuth()) || s.Theme == "" || s.Language == "" || s.Timezone == "" || s.EntryDirection == "" || s.DisplayMode == "" || s.DefaultHomePage == "" {
		return locale.NewLocalizedError("error.settings_mandatory_fields")
	}

	if s.CJKReadingSpeed <= 0 || s.DefaultReadingSpeed <= 0 {
		return locale.NewLocalizedError("error.settings_reading_speed_is_positive")
	}

	if s.Confirmation == "" {
		// Firefox insists on auto-completing the password field.
		// If the confirmation field is blank, the user probably
		// didn't intend to change their password.
		s.Password = ""
	} else if s.Password != "" {
		if s.Password != s.Confirmation {
			return locale.NewLocalizedError("error.different_passwords")
		}
	}

	if s.MediaPlaybackRate < 0.25 || s.MediaPlaybackRate > 4 {
		return locale.NewLocalizedError("error.settings_media_playback_rate_range")
	}

	if s.ExternalFontHosts != "" {
		if !validator.IsValidDomainList(s.ExternalFontHosts) {
			return locale.NewLocalizedError("error.settings_invalid_domain_list")
		}
	}

	return nil
}

// NewSettingsForm returns a new SettingsForm.
func NewSettingsForm(r *http.Request) *SettingsForm {
	entriesPerPage, err := strconv.Atoi(r.FormValue("entries_per_page"))
	if err != nil {
		entriesPerPage = 0
	}
	defaultReadingSpeed, err := strconv.Atoi(r.FormValue("default_reading_speed"))
	if err != nil {
		defaultReadingSpeed = 0
	}
	cjkReadingSpeed, err := strconv.Atoi(r.FormValue("cjk_reading_speed"))
	if err != nil {
		cjkReadingSpeed = 0
	}
	mediaPlaybackRate, err := strconv.ParseFloat(r.FormValue("media_playback_rate"), 64)
	if err != nil {
		mediaPlaybackRate = 1
	}
	return &SettingsForm{
		Username:                  r.FormValue("username"),
		Password:                  r.FormValue("password"),
		Confirmation:              r.FormValue("confirmation"),
		Theme:                     r.FormValue("theme"),
		Language:                  r.FormValue("language"),
		Timezone:                  r.FormValue("timezone"),
		EntryDirection:            r.FormValue("entry_direction"),
		EntryOrder:                r.FormValue("entry_order"),
		EntriesPerPage:            int(entriesPerPage),
		KeyboardShortcuts:         r.FormValue("keyboard_shortcuts") == "1",
		ShowReadingTime:           r.FormValue("show_reading_time") == "1",
		CustomCSS:                 r.FormValue("custom_css"),
		CustomJS:                  r.FormValue("custom_js"),
		ExternalFontHosts:         r.FormValue("external_font_hosts"),
		EntrySwipe:                r.FormValue("entry_swipe") == "1",
		GestureNav:                r.FormValue("gesture_nav"),
		DisplayMode:               r.FormValue("display_mode"),
		DefaultReadingSpeed:       int(defaultReadingSpeed),
		CJKReadingSpeed:           int(cjkReadingSpeed),
		DefaultHomePage:           r.FormValue("default_home_page"),
		CategoriesSortingOrder:    r.FormValue("categories_sorting_order"),
		MarkReadOnView:            r.FormValue("mark_read_on_view") == "1",
		MarkReadBehavior:          model.MarkReadBehavior(r.FormValue("mark_read_behavior")),
		MediaPlaybackRate:         mediaPlaybackRate,
		BlockFilterEntryRules:     r.FormValue("block_filter_entry_rules"),
		KeepFilterEntryRules:      r.FormValue("keep_filter_entry_rules"),
		AlwaysOpenExternalLinks:   r.FormValue("always_open_external_links") == "1",
		OpenExternalLinksInNewTab: r.FormValue("open_external_links_in_new_tab") == "1",
	}
}
