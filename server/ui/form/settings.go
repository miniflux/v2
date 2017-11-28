// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form

import (
	"net/http"

	"github.com/miniflux/miniflux2/errors"
	"github.com/miniflux/miniflux2/model"
)

// SettingsForm represents the settings form.
type SettingsForm struct {
	Username     string
	Password     string
	Confirmation string
	Theme        string
	Language     string
	Timezone     string
}

// Merge updates the fields of the given user.
func (s *SettingsForm) Merge(user *model.User) *model.User {
	user.Username = s.Username
	user.Theme = s.Theme
	user.Language = s.Language
	user.Timezone = s.Timezone

	if s.Password != "" {
		user.Password = s.Password
	}

	return user
}

// Validate makes sure the form values are valid.
func (s *SettingsForm) Validate() error {
	if s.Username == "" || s.Theme == "" || s.Language == "" || s.Timezone == "" {
		return errors.NewLocalizedError("The username, theme, language and timezone fields are mandatory.")
	}

	if s.Password != "" {
		if s.Password != s.Confirmation {
			return errors.NewLocalizedError("Passwords are not the same.")
		}

		if len(s.Password) < 6 {
			return errors.NewLocalizedError("You must use at least 6 characters")
		}
	}

	return nil
}

// NewSettingsForm returns a new SettingsForm.
func NewSettingsForm(r *http.Request) *SettingsForm {
	return &SettingsForm{
		Username:     r.FormValue("username"),
		Password:     r.FormValue("password"),
		Confirmation: r.FormValue("confirmation"),
		Theme:        r.FormValue("theme"),
		Language:     r.FormValue("language"),
		Timezone:     r.FormValue("timezone"),
	}
}
