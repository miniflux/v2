// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// +build integration

package tests

import (
	"testing"

	miniflux "miniflux.app/client"
)

func TestWithWrongCredentials(t *testing.T) {
	client := miniflux.New(testBaseURL, "invalid", "invalid")
	_, err := client.Users()
	if err == nil {
		t.Fatal(`Using bad credentials should raise an error`)
	}

	if err != miniflux.ErrNotAuthorized {
		t.Fatal(`A "Not Authorized" error should be raised`)
	}
}

func TestGetCurrentLoggedUser(t *testing.T) {
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.Me()
	if err != nil {
		t.Fatal(err)
	}

	if user.ID == 0 {
		t.Fatalf(`Invalid userID, got %q`, user.ID)
	}

	if user.Username != testAdminUsername {
		t.Fatalf(`Invalid username, got %q`, user.Username)
	}
}

func TestGetUsers(t *testing.T) {
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	users, err := client.Users()
	if err != nil {
		t.Fatal(err)
	}

	if len(users) == 0 {
		t.Fatal("The list of users is empty")
	}

	if users[0].ID == 0 {
		t.Fatalf(`Invalid userID, got "%v"`, users[0].ID)
	}

	if users[0].Username != testAdminUsername {
		t.Fatalf(`Invalid username, got "%v" instead of "%v"`, users[0].Username, testAdminUsername)
	}

	if users[0].Password != "" {
		t.Fatalf(`Invalid password, got "%v"`, users[0].Password)
	}

	if users[0].Language != "en_US" {
		t.Fatalf(`Invalid language, got "%v"`, users[0].Language)
	}

	if users[0].Theme != "default" {
		t.Fatalf(`Invalid theme, got "%v"`, users[0].Theme)
	}

	if users[0].Timezone != "UTC" {
		t.Fatalf(`Invalid timezone, got "%v"`, users[0].Timezone)
	}

	if !users[0].IsAdmin {
		t.Fatalf(`Invalid role, got "%v"`, users[0].IsAdmin)
	}
}

func TestCreateStandardUser(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	if user.ID == 0 {
		t.Fatalf(`Invalid userID, got "%v"`, user.ID)
	}

	if user.Username != username {
		t.Fatalf(`Invalid username, got "%v" instead of "%v"`, user.Username, username)
	}

	if user.Password != "" {
		t.Fatalf(`Invalid password, got "%v"`, user.Password)
	}

	if user.Language != "en_US" {
		t.Fatalf(`Invalid language, got "%v"`, user.Language)
	}

	if user.Theme != "default" {
		t.Fatalf(`Invalid theme, got "%v"`, user.Theme)
	}

	if user.Timezone != "UTC" {
		t.Fatalf(`Invalid timezone, got "%v"`, user.Timezone)
	}

	if user.IsAdmin {
		t.Fatalf(`Invalid role, got "%v"`, user.IsAdmin)
	}

	if user.LastLoginAt != nil {
		t.Fatalf(`Invalid last login date, got "%v"`, user.LastLoginAt)
	}
}

func TestRemoveUser(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.DeleteUser(user.ID); err != nil {
		t.Fatalf(`Unable to remove user: "%v"`, err)
	}
}

func TestGetUserByID(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.UserByID(99999)
	if err == nil {
		t.Fatal(`Should returns a 404`)
	}

	user, err = client.UserByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if user.ID == 0 {
		t.Fatalf(`Invalid userID, got "%v"`, user.ID)
	}

	if user.Username != username {
		t.Fatalf(`Invalid username, got "%v" instead of "%v"`, user.Username, username)
	}

	if user.Password != "" {
		t.Fatalf(`Invalid password, got "%v"`, user.Password)
	}

	if user.Language != "en_US" {
		t.Fatalf(`Invalid language, got "%v"`, user.Language)
	}

	if user.Theme != "default" {
		t.Fatalf(`Invalid theme, got "%v"`, user.Theme)
	}

	if user.Timezone != "UTC" {
		t.Fatalf(`Invalid timezone, got "%v"`, user.Timezone)
	}

	if user.IsAdmin {
		t.Fatalf(`Invalid role, got "%v"`, user.IsAdmin)
	}

	if user.LastLoginAt != nil {
		t.Fatalf(`Invalid last login date, got "%v"`, user.LastLoginAt)
	}
}

func TestGetUserByUsername(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.UserByUsername("missinguser")
	if err == nil {
		t.Fatal(`Should returns a 404`)
	}

	user, err = client.UserByUsername(username)
	if err != nil {
		t.Fatal(err)
	}

	if user.ID == 0 {
		t.Fatalf(`Invalid userID, got "%v"`, user.ID)
	}

	if user.Username != username {
		t.Fatalf(`Invalid username, got "%v" instead of "%v"`, user.Username, username)
	}

	if user.Password != "" {
		t.Fatalf(`Invalid password, got "%v"`, user.Password)
	}

	if user.Language != "en_US" {
		t.Fatalf(`Invalid language, got "%v"`, user.Language)
	}

	if user.Theme != "default" {
		t.Fatalf(`Invalid theme, got "%v"`, user.Theme)
	}

	if user.Timezone != "UTC" {
		t.Fatalf(`Invalid timezone, got "%v"`, user.Timezone)
	}

	if user.IsAdmin {
		t.Fatalf(`Invalid role, got "%v"`, user.IsAdmin)
	}

	if user.LastLoginAt != nil {
		t.Fatalf(`Invalid last login date, got "%v"`, user.LastLoginAt)
	}
}

func TestUpdateUserTheme(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	theme := "black"
	user, err = client.UpdateUser(user.ID, &miniflux.UserModification{Theme: &theme})
	if err != nil {
		t.Fatal(err)
	}

	if user.Theme != theme {
		t.Fatalf(`Unable to update user Theme: got "%v" instead of "%v"`, user.Theme, theme)
	}
}

func TestUpdateUserThemeWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	theme := "something that doesn't exists"
	_, err = client.UpdateUser(user.ID, &miniflux.UserModification{Theme: &theme})
	if err == nil {
		t.Fatal(`Updating a user Theme with an invalid value should raise an error`)
	}
}

func TestCannotCreateDuplicateUser(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateUser(username, testStandardPassword, false)
	if err == nil {
		t.Fatal(`Duplicate users should not be allowed`)
	}
}

func TestCannotListUsersAsNonAdmin(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.New(testBaseURL, username, testStandardPassword)
	_, err = client.Users()
	if err == nil {
		t.Fatal(`Standard users should not be able to list any users`)
	}

	if err != miniflux.ErrForbidden {
		t.Fatal(`A "Forbidden" error should be raised`)
	}
}

func TestCannotGetUserAsNonAdmin(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.New(testBaseURL, username, testStandardPassword)
	_, err = client.UserByID(user.ID)
	if err == nil {
		t.Fatal(`Standard users should not be able to get any users`)
	}

	if err != miniflux.ErrForbidden {
		t.Fatal(`A "Forbidden" error should be raised`)
	}
}

func TestCannotUpdateUserAsNonAdmin(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.New(testBaseURL, username, testStandardPassword)
	_, err = client.UpdateUser(user.ID, &miniflux.UserModification{})
	if err == nil {
		t.Fatal(`Standard users should not be able to update any users`)
	}

	if err != miniflux.ErrForbidden {
		t.Fatal(`A "Forbidden" error should be raised`)
	}
}

func TestCannotCreateUserAsNonAdmin(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.New(testBaseURL, username, testStandardPassword)
	_, err = client.CreateUser(username, testStandardPassword, false)
	if err == nil {
		t.Fatal(`Standard users should not be able to create users`)
	}

	if err != miniflux.ErrForbidden {
		t.Fatal(`A "Forbidden" error should be raised`)
	}
}

func TestCannotDeleteUserAsNonAdmin(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.New(testBaseURL, username, testStandardPassword)
	err = client.DeleteUser(user.ID)
	if err == nil {
		t.Fatal(`Standard users should not be able to remove any users`)
	}

	if err != miniflux.ErrForbidden {
		t.Fatal(`A "Forbidden" error should be raised`)
	}
}
