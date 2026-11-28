// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestGetMeEndpoint(t *testing.T) {
	t.Parallel()

	client := newAdminClient(t)
	user, err := client.Me()
	if err != nil {
		t.Fatal(err)
	}

	if user.Username != adminUsername {
		t.Fatalf(`Invalid username, got %q instead of %q`, user.Username, adminUsername)
	}
}

func TestGetUsersEndpointAsAdmin(t *testing.T) {
	t.Parallel()

	client := newAdminClient(t)
	users, err := client.Users()
	if err != nil {
		t.Fatal(err)
	}

	if len(users) == 0 {
		t.Fatal(`Users should not be empty`)
	}

	if users[0].ID == 0 {
		t.Fatalf(`Invalid userID, got "%v"`, users[0].ID)
	}

	if users[0].Username != adminUsername {
		t.Fatalf(`Invalid username, got "%v" instead of "%v"`, users[0].Username, adminUsername)
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

func TestGetUsersEndpointAsRegularUser(t *testing.T) {
	t.Parallel()

	regularUserClient, _ := newRegularUser(t)
	_, err := regularUserClient.Users()
	if !errors.Is(err, miniflux.ErrForbidden) {
		t.Fatalf(`Regular users should not have access to the users endpoint, got %v`, err)
	}
}

func TestCreateUserEndpointAsAdmin(t *testing.T) {
	t.Parallel()

	client := newAdminClient(t)

	username := randomUsername()
	regularTestUser, err := client.CreateUser(username, regularUserPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	deleteUserOnCleanup(t, regularTestUser.ID)

	if regularTestUser.Username != username {
		t.Fatalf(`Invalid username, got "%v" instead of "%v"`, regularTestUser.Username, username)
	}

	if regularTestUser.Password != "" {
		t.Fatalf(`Invalid password, got "%v"`, regularTestUser.Password)
	}

	if regularTestUser.Language != "en_US" {
		t.Fatalf(`Invalid language, got "%v"`, regularTestUser.Language)
	}

	if regularTestUser.Theme != "light_serif" {
		t.Fatalf(`Invalid theme, got "%v"`, regularTestUser.Theme)
	}

	if regularTestUser.Timezone != "UTC" {
		t.Fatalf(`Invalid timezone, got "%v"`, regularTestUser.Timezone)
	}

	if regularTestUser.IsAdmin {
		t.Fatalf(`Invalid role, got "%v"`, regularTestUser.IsAdmin)
	}

	if regularTestUser.EntriesPerPage != 100 {
		t.Fatalf(`Invalid entries per page, got "%v"`, regularTestUser.EntriesPerPage)
	}

	if regularTestUser.DisplayMode != "standalone" {
		t.Fatalf(`Invalid web app display mode, got "%v"`, regularTestUser.DisplayMode)
	}

	if regularTestUser.GestureNav != "tap" {
		t.Fatalf(`Invalid gesture navigation, got "%v"`, regularTestUser.GestureNav)
	}

	if regularTestUser.DefaultReadingSpeed != 265 {
		t.Fatalf(`Invalid default reading speed, got "%v"`, regularTestUser.DefaultReadingSpeed)
	}

	if regularTestUser.CJKReadingSpeed != 500 {
		t.Fatalf(`Invalid cjk reading speed, got "%v"`, regularTestUser.CJKReadingSpeed)
	}
}

func TestCreateUserEndpointAsRegularUser(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)
	_, err := regularUserClient.CreateUser(regularTestUser.Username, regularUserPassword, false)
	if !errors.Is(err, miniflux.ErrForbidden) {
		t.Fatalf(`Regular users should not have access to the create user endpoint, got %v`, err)
	}
}

func TestCannotCreateDuplicateUser(t *testing.T) {
	t.Parallel()

	client := newAdminClient(t)
	_, err := client.CreateUser(adminUsername, adminPassword, true)
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Duplicated users should not be allowed, got %v`, err)
	}
}

func TestRemoveUserEndpointAsAdmin(t *testing.T) {
	t.Parallel()

	client := newAdminClient(t)
	user, err := client.CreateUser(randomUsername(), regularUserPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.DeleteUser(user.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := client.UserByID(user.ID); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`The user should be removed once the delete request returns, got error: %v`, err)
	}
}

func TestRemoveUserEndpointAsRegularUser(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)
	err := regularUserClient.DeleteUser(regularTestUser.ID)
	if !errors.Is(err, miniflux.ErrForbidden) {
		t.Fatalf(`Regular users should not have access to the remove user endpoint, got %v`, err)
	}
}

func TestGetUserByIDEndpointAsAdmin(t *testing.T) {
	t.Parallel()

	client := newAdminClient(t)
	user, err := client.Me()
	if err != nil {
		t.Fatal(err)
	}

	userByID, err := client.UserByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if userByID.ID != user.ID {
		t.Errorf(`Invalid userID, got "%v" instead of "%v"`, userByID.ID, user.ID)
	}

	if userByID.Username != user.Username {
		t.Errorf(`Invalid username, got "%v" instead of "%v"`, userByID.Username, user.Username)
	}

	if userByID.Password != "" {
		t.Errorf(`The password field must be empty, got "%v"`, userByID.Password)
	}

	if userByID.Language != user.Language {
		t.Errorf(`Invalid language, got "%v"`, userByID.Language)
	}

	if userByID.Theme != user.Theme {
		t.Errorf(`Invalid theme, got "%v"`, userByID.Theme)
	}

	if userByID.Timezone != user.Timezone {
		t.Errorf(`Invalid timezone, got "%v"`, userByID.Timezone)
	}

	if userByID.IsAdmin != user.IsAdmin {
		t.Errorf(`Invalid role, got "%v"`, userByID.IsAdmin)
	}

	if userByID.EntriesPerPage != user.EntriesPerPage {
		t.Errorf(`Invalid entries per page, got "%v"`, userByID.EntriesPerPage)
	}

	if userByID.DisplayMode != user.DisplayMode {
		t.Errorf(`Invalid web app display mode, got "%v"`, userByID.DisplayMode)
	}

	if userByID.GestureNav != user.GestureNav {
		t.Errorf(`Invalid gesture navigation, got "%v"`, userByID.GestureNav)
	}

	if userByID.DefaultReadingSpeed != user.DefaultReadingSpeed {
		t.Errorf(`Invalid default reading speed, got "%v"`, userByID.DefaultReadingSpeed)
	}

	if userByID.CJKReadingSpeed != user.CJKReadingSpeed {
		t.Errorf(`Invalid cjk reading speed, got "%v"`, userByID.CJKReadingSpeed)
	}

	if userByID.EntryDirection != user.EntryDirection {
		t.Errorf(`Invalid entry direction, got "%v"`, userByID.EntryDirection)
	}

	if userByID.EntryOrder != user.EntryOrder {
		t.Errorf(`Invalid entry order, got "%v"`, userByID.EntryOrder)
	}
}

func TestGetUserByIDEndpointAsRegularUser(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)
	_, err := regularUserClient.UserByID(regularTestUser.ID)
	if !errors.Is(err, miniflux.ErrForbidden) {
		t.Fatalf(`Regular users should not have access to the user by ID endpoint, got %v`, err)
	}
}

func TestGetUserByUsernameEndpointAsAdmin(t *testing.T) {
	t.Parallel()

	client := newAdminClient(t)
	user, err := client.Me()
	if err != nil {
		t.Fatal(err)
	}

	userByUsername, err := client.UserByUsername(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	if userByUsername.ID != user.ID {
		t.Errorf(`Invalid userID, got "%v" instead of "%v"`, userByUsername.ID, user.ID)
	}

	if userByUsername.Username != user.Username {
		t.Errorf(`Invalid username, got "%v" instead of "%v"`, userByUsername.Username, user.Username)
	}

	if userByUsername.Password != "" {
		t.Errorf(`The password field must be empty, got "%v"`, userByUsername.Password)
	}

	if userByUsername.Language != user.Language {
		t.Errorf(`Invalid language, got "%v"`, userByUsername.Language)
	}

	if userByUsername.Theme != user.Theme {
		t.Errorf(`Invalid theme, got "%v"`, userByUsername.Theme)
	}

	if userByUsername.Timezone != user.Timezone {
		t.Errorf(`Invalid timezone, got "%v"`, userByUsername.Timezone)
	}

	if userByUsername.IsAdmin != user.IsAdmin {
		t.Errorf(`Invalid role, got "%v"`, userByUsername.IsAdmin)
	}

	if userByUsername.EntriesPerPage != user.EntriesPerPage {
		t.Errorf(`Invalid entries per page, got "%v"`, userByUsername.EntriesPerPage)
	}

	if userByUsername.DisplayMode != user.DisplayMode {
		t.Errorf(`Invalid web app display mode, got "%v"`, userByUsername.DisplayMode)
	}

	if userByUsername.GestureNav != user.GestureNav {
		t.Errorf(`Invalid gesture navigation, got "%v"`, userByUsername.GestureNav)
	}

	if userByUsername.DefaultReadingSpeed != user.DefaultReadingSpeed {
		t.Errorf(`Invalid default reading speed, got "%v"`, userByUsername.DefaultReadingSpeed)
	}

	if userByUsername.CJKReadingSpeed != user.CJKReadingSpeed {
		t.Errorf(`Invalid cjk reading speed, got "%v"`, userByUsername.CJKReadingSpeed)
	}

	if userByUsername.EntryDirection != user.EntryDirection {
		t.Errorf(`Invalid entry direction, got "%v"`, userByUsername.EntryDirection)
	}
}

func TestGetUserByUsernameEndpointAsRegularUser(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)
	_, err := regularUserClient.UserByUsername(regularTestUser.Username)
	if !errors.Is(err, miniflux.ErrForbidden) {
		t.Fatalf(`Regular users should not have access to the user by username endpoint, got %v`, err)
	}
}

func TestUpdateUserEndpointByChangingDefaultTheme(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)

	userUpdateRequest := &miniflux.UserModificationRequest{
		Theme: new("dark_serif"),
	}

	updatedUser, err := regularUserClient.UpdateUser(regularTestUser.ID, userUpdateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if updatedUser.Theme != "dark_serif" {
		t.Fatalf(`Invalid theme, got "%v"`, updatedUser.Theme)
	}
}

func TestUpdateUserEndpointByChangingExternalFonts(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)

	userUpdateRequest := &miniflux.UserModificationRequest{
		ExternalFontHosts: new("  fonts.example.org  "),
	}

	updatedUser, err := regularUserClient.UpdateUser(regularTestUser.ID, userUpdateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if updatedUser.ExternalFontHosts != "fonts.example.org" {
		t.Fatalf(`Invalid external font hosts, got "%v"`, updatedUser.ExternalFontHosts)
	}
}

func TestUpdateUserEndpointByChangingExternalFontsWithInvalidValue(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)

	userUpdateRequest := &miniflux.UserModificationRequest{
		ExternalFontHosts: new("'self' *"),
	}

	if _, err := regularUserClient.UpdateUser(regularTestUser.ID, userUpdateRequest); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Updating the user with an invalid external font host should raise an error, got %v`, err)
	}
}

func TestUpdateUserEndpointByChangingCustomJS(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)

	userUpdateRequest := &miniflux.UserModificationRequest{
		CustomJS: new("alert('Hello, World!');"),
	}

	updatedUser, err := regularUserClient.UpdateUser(regularTestUser.ID, userUpdateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if updatedUser.CustomJS != "alert('Hello, World!');" {
		t.Fatalf(`Invalid custom JS, got %q`, updatedUser.CustomJS)
	}
}

func TestUpdateUserEndpointByChangingDefaultThemeToInvalidValue(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)

	userUpdateRequest := &miniflux.UserModificationRequest{
		Theme: new("invalid_theme"),
	}

	_, err := regularUserClient.UpdateUser(regularTestUser.ID, userUpdateRequest)
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Updating the user with an invalid theme should raise an error, got %v`, err)
	}
}

func TestRegularUsersCannotUpdateOtherUsers(t *testing.T) {
	t.Parallel()

	adminClient := newAdminClient(t)
	adminUser, err := adminClient.Me()
	if err != nil {
		t.Fatal(err)
	}

	regularUserClient, _ := newRegularUser(t)

	userUpdateRequest := &miniflux.UserModificationRequest{
		Theme: new("dark_serif"),
	}

	_, err = regularUserClient.UpdateUser(adminUser.ID, userUpdateRequest)
	if !errors.Is(err, miniflux.ErrForbidden) {
		t.Fatalf(`Regular users should not be able to update other users, got %v`, err)
	}
}

func TestMarkUserAsReadEndpoint(t *testing.T) {
	t.Parallel()

	fixture := servertest.NewFeedFixture(t)

	regularUserClient, regularTestUser := newRegularUser(t)
	feedID := fixture.Subscribe(t, regularUserClient)

	if err := regularUserClient.MarkAllAsRead(regularTestUser.ID); err != nil {
		t.Fatal(err)
	}

	results := listFeedEntries(t, regularUserClient, feedID, nil)

	for _, entry := range results.Entries {
		if entry.Status != miniflux.EntryStatusRead {
			t.Errorf(`Status for entry %d was %q instead of %q`, entry.ID, entry.Status, miniflux.EntryStatusRead)
		}
	}
}

func TestCannotMarkUserAsReadAsOtherUser(t *testing.T) {
	t.Parallel()

	adminClient := newAdminClient(t)

	adminUser, err := adminClient.Me()
	if err != nil {
		t.Fatal(err)
	}

	regularUserClient, _ := newRegularUser(t)
	if err := regularUserClient.MarkAllAsRead(adminUser.ID); !errors.Is(err, miniflux.ErrForbidden) {
		t.Fatalf(`Non-admin users should not be able to mark another user as read, got %v`, err)
	}
}

func TestFetchIntegrationsStatusEndpoint(t *testing.T) {
	t.Parallel()

	regularUserClient, _ := newRegularUser(t)

	hasIntegrations, err := regularUserClient.IntegrationsStatus()
	if err != nil {
		t.Fatalf("Failed to fetch integrations status: %v", err)
	}

	if hasIntegrations {
		t.Fatalf("New user should not have integrations configured")
	}
}
