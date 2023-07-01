// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build integration
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

	if users[0].Theme != "light_serif" {
		t.Fatalf(`Invalid theme, got "%v"`, users[0].Theme)
	}

	if users[0].Timezone != "UTC" {
		t.Fatalf(`Invalid timezone, got "%v"`, users[0].Timezone)
	}

	if !users[0].IsAdmin {
		t.Fatalf(`Invalid role, got "%v"`, users[0].IsAdmin)
	}

	if users[0].EntriesPerPage != 100 {
		t.Fatalf(`Invalid entries per page, got "%v"`, users[0].EntriesPerPage)
	}

	if users[0].DisplayMode != "standalone" {
		t.Fatalf(`Invalid web app display mode, got "%v"`, users[0].DisplayMode)
	}

	if users[0].GestureNav != "tap" {
		t.Fatalf(`Invalid gesture navigation, got "%v"`, users[0].GestureNav)
	}

	if users[0].DefaultReadingSpeed != 265 {
		t.Fatalf(`Invalid default reading speed, got "%v"`, users[0].DefaultReadingSpeed)
	}

	if users[0].CJKReadingSpeed != 500 {
		t.Fatalf(`Invalid cjk reading speed, got "%v"`, users[0].CJKReadingSpeed)
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

	if user.Theme != "light_serif" {
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

	if user.EntriesPerPage != 100 {
		t.Fatalf(`Invalid entries per page, got "%v"`, user.EntriesPerPage)
	}

	if user.DisplayMode != "standalone" {
		t.Fatalf(`Invalid web app display mode, got "%v"`, user.DisplayMode)
	}

	if user.DefaultReadingSpeed != 265 {
		t.Fatalf(`Invalid default reading speed, got "%v"`, user.DefaultReadingSpeed)
	}

	if user.CJKReadingSpeed != 500 {
		t.Fatalf(`Invalid cjk reading speed, got "%v"`, user.CJKReadingSpeed)
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

	if user.Theme != "light_serif" {
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

	if user.EntriesPerPage != 100 {
		t.Fatalf(`Invalid entries per page, got "%v"`, user.EntriesPerPage)
	}

	if user.DisplayMode != "standalone" {
		t.Fatalf(`Invalid web app display mode, got "%v"`, user.DisplayMode)
	}

	if user.DefaultReadingSpeed != 265 {
		t.Fatalf(`Invalid default reading speed, got "%v"`, user.DefaultReadingSpeed)
	}

	if user.CJKReadingSpeed != 500 {
		t.Fatalf(`Invalid cjk reading speed, got "%v"`, user.CJKReadingSpeed)
	}
}

func TestGetUserByUsername(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.UserByUsername("missinguser")
	if err == nil {
		t.Fatal(`Should returns a 404`)
	}

	user, err := client.UserByUsername(username)
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

	if user.Theme != "light_serif" {
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

	if user.EntriesPerPage != 100 {
		t.Fatalf(`Invalid entries per page, got "%v"`, user.EntriesPerPage)
	}

	if user.DisplayMode != "standalone" {
		t.Fatalf(`Invalid web app display mode, got "%v"`, user.DisplayMode)
	}

	if user.DefaultReadingSpeed != 265 {
		t.Fatalf(`Invalid default reading speed, got "%v"`, user.DefaultReadingSpeed)
	}

	if user.CJKReadingSpeed != 500 {
		t.Fatalf(`Invalid cjk reading speed, got "%v"`, user.CJKReadingSpeed)
	}
}

func TestUpdateUserTheme(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	theme := "dark_serif"
	user, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{Theme: &theme})
	if err != nil {
		t.Fatal(err)
	}

	if user.Theme != theme {
		t.Fatalf(`Unable to update user Theme: got "%v" instead of "%v"`, user.Theme, theme)
	}
}

func TestUpdateUserFields(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	stylesheet := "body { color: red }"
	swipe := false
	entriesPerPage := 5
	displayMode := "fullscreen"
	defaultReadingSpeed := 380
	cjkReadingSpeed := 200
	user, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{
		Stylesheet:          &stylesheet,
		EntrySwipe:          &swipe,
		EntriesPerPage:      &entriesPerPage,
		DisplayMode:         &displayMode,
		DefaultReadingSpeed: &defaultReadingSpeed,
		CJKReadingSpeed:     &cjkReadingSpeed,
	})
	if err != nil {
		t.Fatal(err)
	}

	if user.Stylesheet != stylesheet {
		t.Fatalf(`Unable to update user stylesheet: got %q instead of %q`, user.Stylesheet, stylesheet)
	}

	if user.EntrySwipe != swipe {
		t.Fatalf(`Unable to update user EntrySwipe: got %v instead of %v`, user.EntrySwipe, swipe)
	}

	if user.EntriesPerPage != entriesPerPage {
		t.Fatalf(`Unable to update user EntriesPerPage: got %q instead of %q`, user.EntriesPerPage, entriesPerPage)
	}

	if user.DisplayMode != displayMode {
		t.Fatalf(`Unable to update user DisplayMode: got %q instead of %q`, user.DisplayMode, displayMode)
	}

	if user.DefaultReadingSpeed != defaultReadingSpeed {
		t.Fatalf(`Invalid default reading speed, got %v instead of %v`, user.DefaultReadingSpeed, defaultReadingSpeed)
	}

	if user.CJKReadingSpeed != cjkReadingSpeed {
		t.Fatalf(`Invalid cjk reading speed, got %v instead of %v`, user.CJKReadingSpeed, cjkReadingSpeed)
	}
}

func TestUpdateUserThemeWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	theme := "invalid"
	_, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{Theme: &theme})
	if err == nil {
		t.Fatal(`Updating a user Theme with an invalid value should raise an error`)
	}
}

func TestUpdateUserLanguageWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	language := "invalid"
	_, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{Language: &language})
	if err == nil {
		t.Fatal(`Updating a user language with an invalid value should raise an error`)
	}
}

func TestUpdateUserTimezoneWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	timezone := "invalid"
	_, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{Timezone: &timezone})
	if err == nil {
		t.Fatal(`Updating a user timezone with an invalid value should raise an error`)
	}
}

func TestUpdateUserEntriesPerPageWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	entriesPerPage := -5
	_, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{EntriesPerPage: &entriesPerPage})
	if err == nil {
		t.Fatal(`Updating a user EntriesPerPage with an invalid value should raise an error`)
	}
}

func TestUpdateUserEntryDirectionWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	entryDirection := "invalid"
	_, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{EntryDirection: &entryDirection})
	if err == nil {
		t.Fatal(`Updating a user EntryDirection with an invalid value should raise an error`)
	}
}

func TestUpdateUserEntryOrderWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	entryOrder := "invalid"
	_, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{EntryOrder: &entryOrder})
	if err == nil {
		t.Fatal(`Updating a user EntryOrder with an invalid value should raise an error`)
	}
}

func TestUpdateUserPasswordWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	password := "short"
	_, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{Password: &password})
	if err == nil {
		t.Fatal(`Updating a user password with an invalid value should raise an error`)
	}
}

func TestUpdateUserDisplayModeWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	displayMode := "invalid"
	_, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{DisplayMode: &displayMode})
	if err == nil {
		t.Fatal(`Updating a user web app display mode with an invalid value should raise an error`)
	}
}

func TestUpdateUserWithEmptyUsernameValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	newUsername := ""
	_, err = client.UpdateUser(user.ID, &miniflux.UserModificationRequest{Username: &newUsername})
	if err == nil {
		t.Fatal(`Updating a user with an empty username should raise an error`)
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
		t.Fatal(`Duplicated users should not be allowed`)
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
	adminClient := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)

	usernameA := getRandomUsername()
	userA, err := adminClient.CreateUser(usernameA, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	usernameB := getRandomUsername()
	_, err = adminClient.CreateUser(usernameB, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	entriesPerPage := 10
	userAClient := miniflux.New(testBaseURL, usernameA, testStandardPassword)
	userAAfterUpdate, err := userAClient.UpdateUser(userA.ID, &miniflux.UserModificationRequest{EntriesPerPage: &entriesPerPage})
	if err != nil {
		t.Fatal(`Standard users should be able to update themselves`)
	}

	if userAAfterUpdate.EntriesPerPage != entriesPerPage {
		t.Fatalf(`The EntriesPerPage field of this user should be updated`)
	}

	isAdmin := true
	_, err = userAClient.UpdateUser(userA.ID, &miniflux.UserModificationRequest{IsAdmin: &isAdmin})
	if err == nil {
		t.Fatal(`Standard users should not be able to become admin`)
	}

	userBClient := miniflux.New(testBaseURL, usernameB, testStandardPassword)
	_, err = userBClient.UpdateUser(userA.ID, &miniflux.UserModificationRequest{})
	if err == nil {
		t.Fatal(`Standard users should not be able to update other users`)
	}

	if err != miniflux.ErrForbidden {
		t.Fatal(`A "Forbidden" error should be raised`)
	}

	stylesheet := "test"
	userC, err := adminClient.UpdateUser(userA.ID, &miniflux.UserModificationRequest{Stylesheet: &stylesheet})
	if err != nil {
		t.Fatal(`Admin users should be able to update any users`)
	}

	if userC.Stylesheet != stylesheet {
		t.Fatalf(`The Stylesheet field of this user should be updated`)
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

func TestMarkUserAsReadAsUser(t *testing.T) {
	username := getRandomUsername()
	adminClient := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := adminClient.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client := miniflux.New(testBaseURL, username, testStandardPassword)
	feed, _ := createFeed(t, client)

	results, err := client.FeedEntries(feed.ID, nil)
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}
	if results.Total == 0 {
		t.Fatalf(`Invalid number of entries: %d`, results.Total)
	}
	if results.Entries[0].Status != miniflux.EntryStatusUnread {
		t.Fatalf(`Invalid entry status, got %q instead of %q`, results.Entries[0].Status, miniflux.EntryStatusUnread)
	}

	if err := client.MarkAllAsRead(user.ID); err != nil {
		t.Fatalf(`Failed to mark user's unread entries as read: %v`, err)
	}

	results, err = client.FeedEntries(feed.ID, nil)
	if err != nil {
		t.Fatalf(`Failed to get updated entries: %v`, err)
	}

	for _, entry := range results.Entries {
		if entry.Status != miniflux.EntryStatusRead {
			t.Errorf(`Status for entry %d was %q instead of %q`, entry.ID, entry.Status, miniflux.EntryStatusRead)
		}
	}
}

func TestCannotMarkUserAsReadAsOtherUser(t *testing.T) {
	username := getRandomUsername()
	adminClient := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	user1, err := adminClient.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	createFeed(t, miniflux.New(testBaseURL, username, testStandardPassword))

	username2 := getRandomUsername()
	if _, err = adminClient.CreateUser(username2, testStandardPassword, false); err != nil {
		t.Fatal(err)
	}

	client := miniflux.New(testBaseURL, username2, testStandardPassword)
	err = client.MarkAllAsRead(user1.ID)
	if err == nil {
		t.Fatalf(`Non-admin users should not be able to mark another user as read`)
	}
	if err != miniflux.ErrForbidden {
		t.Errorf(`A "Forbidden" error should be raised, got %q`, err)
	}
}
