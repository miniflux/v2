// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form

import (
	"errors"
	"net/http"
)

type AuthForm struct {
	Username string
	Password string
}

func (a AuthForm) Validate() error {
	if a.Username == "" || a.Password == "" {
		return errors.New("All fields are mandatory.")
	}

	return nil
}

func NewAuthForm(r *http.Request) *AuthForm {
	return &AuthForm{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}
}
