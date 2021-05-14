// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"
	"strconv"

	"miniflux.app/errors"
	"miniflux.app/model"
)

// SettingsForm represents the settings form.
type SettingsForm struct {
	Username          string
	Password          string
	Confirmation      string
	Theme             string
	Language          string
	Timezone          string
	EntryDirection    string
	EntryOrder        string
	EntriesPerPage    int
	KeyboardShortcuts bool
	ShowReadingTime   bool
	CustomCSS         string
	EntrySwipe        bool
	DisplayMode       string
}

// Merge updates the fields of the given user.
func (s *SettingsForm) Merge(user *model.User) *model.User {
	user.Username = s.Username
	user.Theme = s.Theme
	user.Language = s.Language
	user.Timezone = s.Timezone
	user.EntryDirection = s.EntryDirection
	user.EntryOrder = s.EntryOrder
	user.EntriesPerPage = s.EntriesPerPage
	user.KeyboardShortcuts = s.KeyboardShortcuts
	user.ShowReadingTime = s.ShowReadingTime
	user.Stylesheet = s.CustomCSS
	user.EntrySwipe = s.EntrySwipe
	user.DisplayMode = s.DisplayMode

	if s.Password != "" {
		user.Password = s.Password
	}

	return user
}

// Validate makes sure the form values are valid.
func (s *SettingsForm) Validate() error {
	if s.Username == "" || s.Theme == "" || s.Language == "" || s.Timezone == "" || s.EntryDirection == "" || s.DisplayMode == "" {
		return errors.NewLocalizedError("error.settings_mandatory_fields")
	}

	if s.Confirmation == "" {
		// Firefox insists on auto-completing the password field.
		// If the confirmation field is blank, the user probably
		// didn't intend to change their password.
		s.Password = ""
	} else if s.Password != "" {
		if s.Password != s.Confirmation {
			return errors.NewLocalizedError("error.different_passwords")
		}
	}

	return nil
}

// NewSettingsForm returns a new SettingsForm.
func NewSettingsForm(r *http.Request) *SettingsForm {
	entriesPerPage, err := strconv.ParseInt(r.FormValue("entries_per_page"), 10, 64)
	if err != nil {
		entriesPerPage = 0
	}
	return &SettingsForm{
		Username:          r.FormValue("username"),
		Password:          r.FormValue("password"),
		Confirmation:      r.FormValue("confirmation"),
		Theme:             r.FormValue("theme"),
		Language:          r.FormValue("language"),
		Timezone:          r.FormValue("timezone"),
		EntryDirection:    r.FormValue("entry_direction"),
		EntryOrder:        r.FormValue("entry_order"),
		EntriesPerPage:    int(entriesPerPage),
		KeyboardShortcuts: r.FormValue("keyboard_shortcuts") == "1",
		ShowReadingTime:   r.FormValue("show_reading_time") == "1",
		CustomCSS:         r.FormValue("custom_css"),
		EntrySwipe:        r.FormValue("entry_swipe") == "1",
		DisplayMode:       r.FormValue("display_mode"),
	}
}
