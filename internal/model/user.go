// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"time"

	"miniflux.app/v2/internal/timezone"
)

// User represents a user in the system.
type User struct {
	ID                     int64      `json:"id"`
	Username               string     `json:"username"`
	Password               string     `json:"-"`
	IsAdmin                bool       `json:"is_admin"`
	Theme                  string     `json:"theme"`
	Language               string     `json:"language"`
	Timezone               string     `json:"timezone"`
	EntryDirection         string     `json:"entry_sorting_direction"`
	EntryOrder             string     `json:"entry_sorting_order"`
	Stylesheet             string     `json:"stylesheet"`
	GoogleID               string     `json:"google_id"`
	OpenIDConnectID        string     `json:"openid_connect_id"`
	EntriesPerPage         int        `json:"entries_per_page"`
	KeyboardShortcuts      bool       `json:"keyboard_shortcuts"`
	ShowReadingTime        bool       `json:"show_reading_time"`
	EntrySwipe             bool       `json:"entry_swipe"`
	GestureNav             string     `json:"gesture_nav"`
	LastLoginAt            *time.Time `json:"last_login_at"`
	DisplayMode            string     `json:"display_mode"`
	DefaultReadingSpeed    int        `json:"default_reading_speed"`
	CJKReadingSpeed        int        `json:"cjk_reading_speed"`
	DefaultHomePage        string     `json:"default_home_page"`
	CategoriesSortingOrder string     `json:"categories_sorting_order"`
	MarkReadOnView         bool       `json:"mark_read_on_view"`
}

// UserCreationRequest represents the request to create a user.
type UserCreationRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	IsAdmin         bool   `json:"is_admin"`
	GoogleID        string `json:"google_id"`
	OpenIDConnectID string `json:"openid_connect_id"`
}

// UserModificationRequest represents the request to update a user.
type UserModificationRequest struct {
	Username               *string `json:"username"`
	Password               *string `json:"password"`
	Theme                  *string `json:"theme"`
	Language               *string `json:"language"`
	Timezone               *string `json:"timezone"`
	EntryDirection         *string `json:"entry_sorting_direction"`
	EntryOrder             *string `json:"entry_sorting_order"`
	Stylesheet             *string `json:"stylesheet"`
	GoogleID               *string `json:"google_id"`
	OpenIDConnectID        *string `json:"openid_connect_id"`
	EntriesPerPage         *int    `json:"entries_per_page"`
	IsAdmin                *bool   `json:"is_admin"`
	KeyboardShortcuts      *bool   `json:"keyboard_shortcuts"`
	ShowReadingTime        *bool   `json:"show_reading_time"`
	EntrySwipe             *bool   `json:"entry_swipe"`
	GestureNav             *string `json:"gesture_nav"`
	DisplayMode            *string `json:"display_mode"`
	DefaultReadingSpeed    *int    `json:"default_reading_speed"`
	CJKReadingSpeed        *int    `json:"cjk_reading_speed"`
	DefaultHomePage        *string `json:"default_home_page"`
	CategoriesSortingOrder *string `json:"categories_sorting_order"`
	MarkReadOnView         *bool   `json:"mark_read_on_view"`
}

// Patch updates the User object with the modification request.
func (u *UserModificationRequest) Patch(user *User) {
	if u.Username != nil {
		user.Username = *u.Username
	}

	if u.Password != nil {
		user.Password = *u.Password
	}

	if u.IsAdmin != nil {
		user.IsAdmin = *u.IsAdmin
	}

	if u.Theme != nil {
		user.Theme = *u.Theme
	}

	if u.Language != nil {
		user.Language = *u.Language
	}

	if u.Timezone != nil {
		user.Timezone = *u.Timezone
	}

	if u.EntryDirection != nil {
		user.EntryDirection = *u.EntryDirection
	}

	if u.EntryOrder != nil {
		user.EntryOrder = *u.EntryOrder
	}

	if u.Stylesheet != nil {
		user.Stylesheet = *u.Stylesheet
	}

	if u.GoogleID != nil {
		user.GoogleID = *u.GoogleID
	}

	if u.OpenIDConnectID != nil {
		user.OpenIDConnectID = *u.OpenIDConnectID
	}

	if u.EntriesPerPage != nil {
		user.EntriesPerPage = *u.EntriesPerPage
	}

	if u.KeyboardShortcuts != nil {
		user.KeyboardShortcuts = *u.KeyboardShortcuts
	}

	if u.ShowReadingTime != nil {
		user.ShowReadingTime = *u.ShowReadingTime
	}

	if u.EntrySwipe != nil {
		user.EntrySwipe = *u.EntrySwipe
	}

	if u.GestureNav != nil {
		user.GestureNav = *u.GestureNav
	}

	if u.DisplayMode != nil {
		user.DisplayMode = *u.DisplayMode
	}

	if u.DefaultReadingSpeed != nil {
		user.DefaultReadingSpeed = *u.DefaultReadingSpeed
	}

	if u.CJKReadingSpeed != nil {
		user.CJKReadingSpeed = *u.CJKReadingSpeed
	}

	if u.DefaultHomePage != nil {
		user.DefaultHomePage = *u.DefaultHomePage
	}

	if u.CategoriesSortingOrder != nil {
		user.CategoriesSortingOrder = *u.CategoriesSortingOrder
	}

	if u.MarkReadOnView != nil {
		user.MarkReadOnView = *u.MarkReadOnView
	}
}

// UseTimezone converts last login date to the given timezone.
func (u *User) UseTimezone(tz string) {
	if u.LastLoginAt != nil {
		*u.LastLoginAt = timezone.Convert(tz, *u.LastLoginAt)
	}
}

// Users represents a list of users.
type Users []*User

// UseTimezone converts last login timestamp of all users to the given timezone.
func (u Users) UseTimezone(tz string) {
	for _, user := range u {
		user.UseTimezone(tz)
	}
}
