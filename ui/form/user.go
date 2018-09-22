// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"

	"miniflux.app/errors"
	"miniflux.app/model"
)

// UserForm represents the user form.
type UserForm struct {
	Username     string
	Password     string
	Confirmation string
	IsAdmin      bool
}

// ValidateCreation validates user creation.
func (u UserForm) ValidateCreation() error {
	if u.Username == "" || u.Password == "" || u.Confirmation == "" {
		return errors.NewLocalizedError("error.fields_mandatory")
	}

	if u.Password != u.Confirmation {
		return errors.NewLocalizedError("error.different_passwords")
	}

	if len(u.Password) < 6 {
		return errors.NewLocalizedError("error.password_min_length")
	}

	return nil
}

// ValidateModification validates user modification.
func (u UserForm) ValidateModification() error {
	if u.Username == "" {
		return errors.NewLocalizedError("error.user_mandatory_fields")
	}

	if u.Password != "" {
		if u.Password != u.Confirmation {
			return errors.NewLocalizedError("error.different_passwords")
		}

		if len(u.Password) < 6 {
			return errors.NewLocalizedError("error.password_min_length")
		}
	}

	return nil
}

// ToUser returns a User from the form values.
func (u UserForm) ToUser() *model.User {
	return &model.User{
		Username: u.Username,
		Password: u.Password,
		IsAdmin:  u.IsAdmin,
	}
}

// Merge updates the fields of the given user.
func (u UserForm) Merge(user *model.User) *model.User {
	user.Username = u.Username
	user.IsAdmin = u.IsAdmin

	if u.Password != "" {
		user.Password = u.Password
	}

	return user
}

// NewUserForm returns a new UserForm.
func NewUserForm(r *http.Request) *UserForm {
	return &UserForm{
		Username:     r.FormValue("username"),
		Password:     r.FormValue("password"),
		Confirmation: r.FormValue("confirmation"),
		IsAdmin:      r.FormValue("is_admin") == "1",
	}
}
