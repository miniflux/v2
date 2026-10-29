// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"

	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
)

// AuthSettingsForm represents the authentication settings form.
type AuthSettingsForm struct {
	Username     string
	Password     string
	Confirmation string
}

// Merge updates the fields of the given user.
func (a *AuthSettingsForm) Merge(user *model.User) *model.User {
	user.Username = a.Username

	if a.Password != "" {
		user.Password = a.Password
	}

	return user
}

// Validate makes sure the form values are valid.
func (a *AuthSettingsForm) Validate() *locale.LocalizedError {
	if a.Username == "" {
		return locale.NewLocalizedError("error.settings_mandatory_fields")
	}

	if a.Confirmation == "" {
		// Firefox insists on auto-completing the password field.
		// If the confirmation field is blank, the user probably
		// didn't intend to change their password.
		a.Password = ""
	} else if a.Password != "" {
		if a.Password != a.Confirmation {
			return locale.NewLocalizedError("error.different_passwords")
		}
	}

	return nil
}

// NewAuthSettingsForm returns a new AuthSettingsForm.
func NewAuthSettingsForm(r *http.Request) *AuthSettingsForm {
	return &AuthSettingsForm{
		Username:     r.FormValue("username"),
		Password:     r.FormValue("password"),
		Confirmation: r.FormValue("confirmation"),
	}
}
