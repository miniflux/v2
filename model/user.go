// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import (
	"errors"
	"time"
)

// User represents a user in the system.
type User struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Password    string     `json:"password,omitempty"`
	IsAdmin     bool       `json:"is_admin"`
	Theme       string     `json:"theme"`
	Language    string     `json:"language"`
	Timezone    string     `json:"timezone"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

func (u User) ValidateUserCreation() error {
	if err := u.ValidateUserLogin(); err != nil {
		return err
	}

	if err := u.ValidatePassword(); err != nil {
		return err
	}

	return nil
}

func (u User) ValidateUserModification() error {
	if u.Username == "" {
		return errors.New("The username is mandatory")
	}

	if err := u.ValidatePassword(); err != nil {
		return err
	}

	return nil
}

func (u User) ValidateUserLogin() error {
	if u.Username == "" {
		return errors.New("The username is mandatory")
	}

	if u.Password == "" {
		return errors.New("The password is mandatory")
	}

	return nil
}

func (u User) ValidatePassword() error {
	if u.Password != "" && len(u.Password) < 6 {
		return errors.New("The password must have at least 6 characters")
	}

	return nil
}

// Merge update the current user with another user.
func (u *User) Merge(override *User) {
	if u.Username != override.Username {
		u.Username = override.Username
	}

	if u.Password != override.Password {
		u.Password = override.Password
	}

	if u.IsAdmin != override.IsAdmin {
		u.IsAdmin = override.IsAdmin
	}

	if u.Theme != override.Theme {
		u.Theme = override.Theme
	}

	if u.Language != override.Language {
		u.Language = override.Language
	}

	if u.Timezone != override.Timezone {
		u.Timezone = override.Timezone
	}
}

// Users represents a list of users.
type Users []*User
