// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form

import (
	"errors"
	"github.com/miniflux/miniflux2/model"
	"net/http"
)

type UserForm struct {
	Username     string
	Password     string
	Confirmation string
	IsAdmin      bool
}

func (u UserForm) ValidateCreation() error {
	if u.Username == "" || u.Password == "" || u.Confirmation == "" {
		return errors.New("All fields are mandatory.")
	}

	if u.Password != u.Confirmation {
		return errors.New("Passwords are not the same.")
	}

	if len(u.Password) < 6 {
		return errors.New("You must use at least 6 characters.")
	}

	return nil
}

func (u UserForm) ValidateModification() error {
	if u.Username == "" {
		return errors.New("The username is mandatory.")
	}

	if u.Password != "" {
		if u.Password != u.Confirmation {
			return errors.New("Passwords are not the same.")
		}

		if len(u.Password) < 6 {
			return errors.New("You must use at least 6 characters.")
		}
	}

	return nil
}

func (u UserForm) ToUser() *model.User {
	return &model.User{
		Username: u.Username,
		Password: u.Password,
		IsAdmin:  u.IsAdmin,
	}
}

func (u UserForm) Merge(user *model.User) *model.User {
	user.Username = u.Username
	user.IsAdmin = u.IsAdmin

	if u.Password != "" {
		user.Password = u.Password
	}

	return user
}

func NewUserForm(r *http.Request) *UserForm {
	return &UserForm{
		Username:     r.FormValue("username"),
		Password:     r.FormValue("password"),
		Confirmation: r.FormValue("confirmation"),
		IsAdmin:      r.FormValue("is_admin") == "1",
	}
}
