// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"

	"miniflux.app/v2/internal/errors"
)

// AuthForm represents the authentication form.
type AuthForm struct {
	Username string
	Password string
}

// Validate makes sure the form values are valid.
func (a AuthForm) Validate() error {
	if a.Username == "" || a.Password == "" {
		return errors.NewLocalizedError("error.fields_mandatory")
	}

	return nil
}

// NewAuthForm returns a new AuthForm.
func NewAuthForm(r *http.Request) *AuthForm {
	return &AuthForm{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}
}
