// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"

	miniflux "miniflux.app/v2/client"
)

const skipIntegrationTestsMessage = `Set TEST_MINIFLUX_* environment variables to run the API integration tests`

type integrationTestConfig struct {
	testBaseURL           string
	testAdminUsername     string
	testAdminPassword     string
	testRegularUsername   string
	testRegularPassword   string
	testFeedURL           string
	testFeedTitle         string
	testSubscriptionTitle string
	testWebsiteURL        string
}

func newIntegrationTestConfig() *integrationTestConfig {
	getDefaultEnvValues := func(key, defaultValue string) string {
		value := os.Getenv(key)
		if value == "" {
			return defaultValue
		}
		return value
	}

	return &integrationTestConfig{
		testBaseURL:           getDefaultEnvValues("TEST_MINIFLUX_BASE_URL", ""),
		testAdminUsername:     getDefaultEnvValues("TEST_MINIFLUX_ADMIN_USERNAME", ""),
		testAdminPassword:     getDefaultEnvValues("TEST_MINIFLUX_ADMIN_PASSWORD", ""),
		testRegularUsername:   getDefaultEnvValues("TEST_MINIFLUX_REGULAR_USERNAME_PREFIX", "regular_test_user"),
		testRegularPassword:   getDefaultEnvValues("TEST_MINIFLUX_REGULAR_PASSWORD", "regular_test_user_password"),
		testFeedURL:           getDefaultEnvValues("TEST_MINIFLUX_FEED_URL", "https://miniflux.app/feed.xml"),
		testFeedTitle:         getDefaultEnvValues("TEST_MINIFLUX_FEED_TITLE", "Miniflux"),
		testSubscriptionTitle: getDefaultEnvValues("TEST_MINIFLUX_SUBSCRIPTION_TITLE", "Miniflux Releases"),
		testWebsiteURL:        getDefaultEnvValues("TEST_MINIFLUX_WEBSITE_URL", "https://miniflux.app"),
	}
}

func (c *integrationTestConfig) isConfigured() bool {
	return c.testBaseURL != "" && c.testAdminUsername != "" && c.testAdminPassword != "" && c.testFeedURL != "" && c.testFeedTitle != "" && c.testSubscriptionTitle != "" && c.testWebsiteURL != ""
}

func (c *integrationTestConfig) genRandomUsername() string {
	return fmt.Sprintf("%s_%10d", c.testRegularUsername, rand.Int())
}

func TestIncorrectEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient("incorrect url")
	_, err := client.Users()
	if err == nil {
		t.Fatal(`Using an incorrect URL should raise an error`)
	}
}

func TestHealthcheckEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL)
	if err := client.Healthcheck(); err != nil {
		t.Fatal(err)
	}
}

func TestVersionEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	version, err := client.Version()
	if err != nil {
		t.Fatal(err)
	}

	if version.Version == "" {
		t.Fatal(`Version should not be empty`)
	}

	if version.Commit == "" {
		t.Fatal(`Commit should not be empty`)
	}

	if version.BuildDate == "" {
		t.Fatal(`Build date should not be empty`)
	}

	if version.GoVersion == "" {
		t.Fatal(`Go version should not be empty`)
	}

	if version.Compiler == "" {
		t.Fatal(`Compiler should not be empty`)
	}

	if version.Arch == "" {
		t.Fatal(`Arch should not be empty`)
	}

	if version.OS == "" {
		t.Fatal(`OS should not be empty`)
	}
}

func TestInvalidCredentials(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, "invalid", "invalid")
	_, err := client.Users()
	if err == nil {
		t.Fatal(`Using bad credentials should raise an error`)
	}

	if err != miniflux.ErrNotAuthorized {
		t.Fatal(`A "Not Authorized" error should be raised`)
	}
}

func TestGetMeEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	user, err := client.Me()
	if err != nil {
		t.Fatal(err)
	}

	if user.Username != testConfig.testAdminUsername {
		t.Fatalf(`Invalid username, got %q instead of %q`, user.Username, testConfig.testAdminUsername)
	}
}

func TestGetUsersEndpointAsAdmin(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
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

	if users[0].Username != testConfig.testAdminUsername {
		t.Fatalf(`Invalid username, got "%v" instead of "%v"`, users[0].Username, testConfig.testAdminUsername)
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
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	_, err = regularUserClient.Users()
	if err == nil {
		t.Fatal(`Regular users should not have access to the users endpoint`)
	}
}

func TestCreateUserEndpointAsAdmin(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	username := testConfig.genRandomUsername()
	regularTestUser, err := client.CreateUser(username, testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeleteUser(regularTestUser.ID)

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
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	_, err = regularUserClient.CreateUser(regularTestUser.Username, testConfig.testRegularPassword, false)
	if err == nil {
		t.Fatal(`Regular users should not have access to the create user endpoint`)
	}
}

func TestCannotCreateDuplicateUser(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	_, err := client.CreateUser(testConfig.testAdminUsername, testConfig.testAdminPassword, true)
	if err == nil {
		t.Fatal(`Duplicated users should not be allowed`)
	}
}

func TestRemoveUserEndpointAsAdmin(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	user, err := client.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.DeleteUser(user.ID); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveUserEndpointAsRegularUser(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	err = regularUserClient.DeleteUser(regularTestUser.ID)
	if err == nil {
		t.Fatal(`Regular users should not have access to the remove user endpoint`)
	}
}

func TestGetUserByIDEndpointAsAdmin(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
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
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	_, err = regularUserClient.UserByID(regularTestUser.ID)
	if err == nil {
		t.Fatal(`Regular users should not have access to the user by ID endpoint`)
	}
}

func TestGetUserByUsernameEndpointAsAdmin(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
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
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	_, err = regularUserClient.UserByUsername(regularTestUser.Username)
	if err == nil {
		t.Fatal(`Regular users should not have access to the user by username endpoint`)
	}
}

func TestUpdateUserEndpointByChangingDefaultTheme(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	userUpdateRequest := &miniflux.UserModificationRequest{
		Theme: miniflux.SetOptionalField("dark_serif"),
	}

	updatedUser, err := regularUserClient.UpdateUser(regularTestUser.ID, userUpdateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if updatedUser.Theme != "dark_serif" {
		t.Fatalf(`Invalid theme, got "%v"`, updatedUser.Theme)
	}
}

func TestUpdateUserEndpointByChangingDefaultThemeToInvalidValue(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	userUpdateRequest := &miniflux.UserModificationRequest{
		Theme: miniflux.SetOptionalField("invalid_theme"),
	}

	_, err = regularUserClient.UpdateUser(regularTestUser.ID, userUpdateRequest)
	if err == nil {
		t.Fatal(`Updating the user with an invalid theme should raise an error`)
	}
}

func TestRegularUsersCannotUpdateOtherUsers(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	adminUser, err := adminClient.Me()
	if err != nil {
		t.Fatal(err)
	}

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	userUpdateRequest := &miniflux.UserModificationRequest{
		Theme: miniflux.SetOptionalField("dark_serif"),
	}

	_, err = regularUserClient.UpdateUser(adminUser.ID, userUpdateRequest)
	if err == nil {
		t.Fatal(`Regular users should not be able to update other users`)
	}
}

func TestMarkUserAsReadEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.MarkAllAsRead(regularTestUser.ID); err != nil {
		t.Fatal(err)
	}

	results, err := regularUserClient.FeedEntries(feedID, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range results.Entries {
		if entry.Status != miniflux.EntryStatusRead {
			t.Errorf(`Status for entry %d was %q instead of %q`, entry.ID, entry.Status, miniflux.EntryStatusRead)
		}
	}
}

func TestCannotMarkUserAsReadAsOtherUser(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	adminUser, err := adminClient.Me()
	if err != nil {
		t.Fatal(err)
	}

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	if err := regularUserClient.MarkAllAsRead(adminUser.ID); err == nil {
		t.Fatalf(`Non-admin users should not be able to mark another user as read`)
	}
}

func TestCreateCategoryEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	categoryName := "My category"
	category, err := regularUserClient.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	if category.ID == 0 {
		t.Errorf(`Invalid categoryID, got "%v"`, category.ID)
	}

	if category.UserID <= 0 {
		t.Errorf(`Invalid userID, got "%v"`, category.UserID)
	}

	if category.Title != categoryName {
		t.Errorf(`Invalid title, got "%v" instead of "%v"`, category.Title, categoryName)
	}
}

func TestCreateCategoryWithEmptyTitle(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	_, err := client.CreateCategory("")
	if err == nil {
		t.Fatalf(`Creating a category with an empty title should raise an error`)
	}
}

func TestCannotCreateDuplicatedCategory(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	categoryName := "My category"

	if _, err := regularUserClient.CreateCategory(categoryName); err != nil {
		t.Fatal(err)
	}

	if _, err = regularUserClient.CreateCategory(categoryName); err == nil {
		t.Fatalf(`Duplicated categories should not be allowed`)
	}
}

func TestUpdateCatgoryEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	categoryName := "My category"
	category, err := regularUserClient.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	updatedCategory, err := regularUserClient.UpdateCategory(category.ID, "new title")
	if err != nil {
		t.Fatal(err)
	}

	if updatedCategory.ID != category.ID {
		t.Errorf(`Invalid categoryID, got "%v"`, updatedCategory.ID)
	}

	if updatedCategory.UserID != regularTestUser.ID {
		t.Errorf(`Invalid userID, got "%v"`, updatedCategory.UserID)
	}

	if updatedCategory.Title != "new title" {
		t.Errorf(`Invalid title, got "%v" instead of "%v"`, updatedCategory.Title, "new title")
	}
}

func TestUpdateInexistingCategory(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	_, err := client.UpdateCategory(123456789, "new title")
	if err == nil {
		t.Fatalf(`Updating an inexisting category should raise an error`)
	}
}
func TestDeleteCategoryEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	categoryName := "My category"
	category, err := regularUserClient.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.DeleteCategory(category.ID); err != nil {
		t.Fatal(err)
	}
}

func TestCannotDeleteInexistingCategory(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	err := client.DeleteCategory(123456789)
	if err == nil {
		t.Fatalf(`Deleting an inexisting category should raise an error`)
	}
}

func TestCannotDeleteCategoryOfAnotherUser(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	err = adminClient.DeleteCategory(category.ID)
	if err == nil {
		t.Fatalf(`Regular users should not be able to delete categories of other users`)
	}
}

func TestGetCategoriesEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	categories, err := regularUserClient.Categories()
	if err != nil {
		t.Fatal(err)
	}

	if len(categories) != 2 {
		t.Fatalf(`Invalid number of categories, got %d instead of %d`, len(categories), 1)
	}

	if categories[0].UserID != regularTestUser.ID {
		t.Fatalf(`Invalid userID, got %d`, categories[0].UserID)
	}

	if categories[0].Title != "All" {
		t.Fatalf(`Invalid title, got %q instead of %q`, categories[0].Title, "All")
	}

	if categories[1].ID != category.ID {
		t.Fatalf(`Invalid categoryID, got %d`, categories[0].ID)
	}

	if categories[1].UserID != regularTestUser.ID {
		t.Fatalf(`Invalid userID, got %d`, categories[0].UserID)
	}

	if categories[1].Title != "My category" {
		t.Fatalf(`Invalid title, got %q instead of %q`, categories[0].Title, "My category")
	}
}

func TestMarkCategoryAsReadEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testConfig.testFeedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.MarkCategoryAsRead(category.ID); err != nil {
		t.Fatal(err)
	}

	results, err := regularUserClient.FeedEntries(feedID, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range results.Entries {
		if entry.Status != miniflux.EntryStatusRead {
			t.Errorf(`Status for entry %d was %q instead of %q`, entry.ID, entry.Status, miniflux.EntryStatusRead)
		}
	}
}

func TestCreateFeedEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)
	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testConfig.testFeedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Errorf(`Invalid feedID, got "%v"`, feedID)
	}
}

func TestCannotCreateDuplicatedFeed(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feedID, got "%v"`, feedID)
	}

	_, err = regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err == nil {
		t.Fatalf(`Duplicated feeds should not be allowed`)
	}
}

func TestCreateFeedWithInexistingCategory(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	_, err = regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testConfig.testFeedURL,
		CategoryID: 123456789,
	})

	if err == nil {
		t.Fatalf(`Creating a feed with an inexisting category should raise an error`)
	}
}

func TestCreateFeedWithEmptyFeedURL(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	_, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: "",
	})
	if err == nil {
		t.Fatalf(`Creating a feed with an empty feed URL should raise an error`)
	}
}

func TestCreateFeedWithInvalidFeedURL(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	_, err := client.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: "invalid_feed_url",
	})
	if err == nil {
		t.Fatalf(`Creating a feed with an invalid feed URL should raise an error`)
	}
}

func TestCreateDisabledFeed(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:  testConfig.testFeedURL,
		Disabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	feed, err := regularUserClient.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if !feed.Disabled {
		t.Fatalf(`The feed should be disabled`)
	}
}

func TestCreateFeedWithDisabledHTTPCache(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:         testConfig.testFeedURL,
		IgnoreHTTPCache: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	feed, err := regularUserClient.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if !feed.IgnoreHTTPCache {
		t.Fatalf(`The feed should ignore the HTTP cache`)
	}
}

func TestCreateFeedWithScraperRule(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:      testConfig.testFeedURL,
		ScraperRules: "article",
	})
	if err != nil {
		t.Fatal(err)
	}

	feed, err := regularUserClient.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if feed.ScraperRules != "article" {
		t.Fatalf(`The feed should have the scraper rules set to "article"`)
	}
}

func TestUpdateFeedEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		FeedURL: miniflux.SetOptionalField("https://example.org/feed.xml"),
	}

	updatedFeed, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if updatedFeed.FeedURL != "https://example.org/feed.xml" {
		t.Fatalf(`Invalid feed URL, got "%v"`, updatedFeed.FeedURL)
	}
}

func TestCannotHaveDuplicateFeedWhenUpdatingFeed(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	if _, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: testConfig.testFeedURL}); err != nil {
		t.Fatal(err)
	}

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: "https://github.com/miniflux/v2/commits.atom",
	})
	if err != nil {
		t.Fatal(err)
	}

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		FeedURL: miniflux.SetOptionalField(testConfig.testFeedURL),
	}

	if _, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest); err == nil {
		t.Fatalf(`Duplicated feeds should not be allowed`)
	}
}

func TestUpdateFeedWithInvalidCategory(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	feedUpdateRequest := &miniflux.FeedModificationRequest{
		CategoryID: miniflux.SetOptionalField(int64(123456789)),
	}

	if _, err := regularUserClient.UpdateFeed(feedID, feedUpdateRequest); err == nil {
		t.Fatalf(`Updating a feed with an inexisting category should raise an error`)
	}
}

func TestMarkFeedAsReadEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.MarkFeedAsRead(feedID); err != nil {
		t.Fatal(err)
	}

	results, err := regularUserClient.FeedEntries(feedID, nil)
	if err != nil {
		t.Fatalf(`Failed to get updated entries: %v`, err)
	}

	for _, entry := range results.Entries {
		if entry.Status != miniflux.EntryStatusRead {
			t.Errorf(`Status for entry %d was %q instead of %q`, entry.ID, entry.Status, miniflux.EntryStatusRead)
		}
	}
}

func TestFetchCountersEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	counters, err := regularUserClient.FetchCounters()
	if err != nil {
		t.Fatal(err)
	}

	if value, ok := counters.ReadCounters[feedID]; ok && value != 0 {
		t.Errorf(`Invalid read counter, got %d`, value)
	}

	if value, ok := counters.UnreadCounters[feedID]; !ok || value == 0 {
		t.Errorf(`Invalid unread counter, got %d`, value)
	}
}

func TestDeleteFeedEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.DeleteFeed(feedID); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshAllFeedsEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	if err := regularUserClient.RefreshAllFeeds(); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshFeedEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.RefreshFeed(feedID); err != nil {
		t.Fatal(err)
	}
}

func TestGetFeedEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	feed, err := regularUserClient.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if feed.ID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, feed.ID)
	}

	if feed.FeedURL != testConfig.testFeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, feed.FeedURL)
	}

	if feed.SiteURL != testConfig.testWebsiteURL {
		t.Fatalf(`Invalid site URL, got %q`, feed.SiteURL)
	}

	if feed.Title != testConfig.testFeedTitle {
		t.Fatalf(`Invalid title, got %q`, feed.Title)
	}
}

func TestGetFeedIcon(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	icon, err := regularUserClient.FeedIcon(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if icon == nil {
		t.Fatalf(`Invalid icon, got nil`)
	}

	if icon.MimeType == "" {
		t.Fatalf(`Invalid mime type, got %q`, icon.MimeType)
	}

	if len(icon.Data) == 0 {
		t.Fatalf(`Invalid data, got empty`)
	}

	icon, err = regularUserClient.Icon(icon.ID)
	if err != nil {
		t.Fatal(err)
	}

	if icon == nil {
		t.Fatalf(`Invalid icon, got nil`)
	}

	if icon.MimeType == "" {
		t.Fatalf(`Invalid mime type, got %q`, icon.MimeType)
	}

	if len(icon.Data) == 0 {
		t.Fatalf(`Invalid data, got empty`)
	}
}

func TestGetFeedIconWithInexistingFeedID(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	_, err := client.FeedIcon(123456789)
	if err == nil {
		t.Fatalf(`Fetching the icon of an inexisting feed should raise an error`)
	}
}

func TestGetFeedsEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	feeds, err := regularUserClient.Feeds()
	if err != nil {
		t.Fatal(err)
	}

	if len(feeds) != 1 {
		t.Fatalf(`Invalid number of feeds, got %d`, len(feeds))
	}

	if feeds[0].ID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, feeds[0].ID)
	}

	if feeds[0].FeedURL != testConfig.testFeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, feeds[0].FeedURL)
	}
}

func TestGetCategoryFeedsEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testConfig.testFeedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	feeds, err := regularUserClient.CategoryFeeds(category.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(feeds) != 1 {
		t.Fatalf(`Invalid number of feeds, got %d`, len(feeds))
	}

	if feeds[0].ID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, feeds[0].ID)
	}

	if feeds[0].FeedURL != testConfig.testFeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, feeds[0].FeedURL)
	}
}

func TestExportEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	if _, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: testConfig.testFeedURL}); err != nil {
		t.Fatal(err)
	}

	exportedData, err := regularUserClient.Export()
	if err != nil {
		t.Fatal(err)
	}

	if len(exportedData) == 0 {
		t.Fatalf(`Invalid exported data, got empty`)
	}

	if !strings.HasPrefix(string(exportedData), "<?xml") {
		t.Fatalf(`Invalid OPML export, got %q`, string(exportedData))
	}
}

func TestImportEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	data := `<?xml version="1.0" encoding="UTF-8"?>
    <opml version="2.0">
        <body>
            <outline text="Test Category">
				<outline title="Test" text="Test" xmlUrl="` + testConfig.testFeedURL + `" htmlUrl="` + testConfig.testWebsiteURL + `"></outline>
			</outline>
		</body>
	</opml>`

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	bytesReader := bytes.NewReader([]byte(data))
	if err := regularUserClient.Import(io.NopCloser(bytesReader)); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverSubscriptionsEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	subscriptions, err := client.Discover(testConfig.testWebsiteURL)
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) == 0 {
		t.Fatalf(`Invalid number of subscriptions, got %d`, len(subscriptions))
	}

	if subscriptions[0].Title != testConfig.testSubscriptionTitle {
		t.Fatalf(`Invalid title, got %q`, subscriptions[0].Title)
	}

	if subscriptions[0].URL != testConfig.testFeedURL {
		t.Fatalf(`Invalid URL, got %q`, subscriptions[0].URL)
	}
}

func TestDiscoverSubscriptionsWithInvalidURL(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	_, err := client.Discover("invalid_url")
	if err == nil {
		t.Fatalf(`Discovering subscriptions with an invalid URL should raise an error`)
	}
}

func TestDiscoverSubscriptionsWithNoSubscription(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	client := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	if _, err := client.Discover(testConfig.testBaseURL); err != miniflux.ErrNotFound {
		t.Fatalf(`Discovering subscriptions with no subscription should raise a 404 error`)
	}
}

func TestGetAllFeedEntriesEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := regularUserClient.FeedEntries(feedID, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Entries) == 0 {
		t.Fatalf(`Invalid number of entries, got %d`, len(results.Entries))
	}

	if results.Total == 0 {
		t.Fatalf(`Invalid total, got %d`, results.Total)
	}

	if results.Entries[0].FeedID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, results.Entries[0].FeedID)
	}

	if results.Entries[0].Feed.FeedURL != testConfig.testFeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, results.Entries[0].Feed.FeedURL)
	}

	if results.Entries[0].Title == "" {
		t.Fatalf(`Invalid title, got empty`)
	}
}

func TestGetAllCategoryEntriesEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	category, err := regularUserClient.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    testConfig.testFeedURL,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := regularUserClient.CategoryEntries(category.ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Entries) == 0 {
		t.Fatalf(`Invalid number of entries, got %d`, len(results.Entries))
	}

	if results.Total == 0 {
		t.Fatalf(`Invalid total, got %d`, results.Total)
	}

	if results.Entries[0].FeedID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, results.Entries[0].FeedID)
	}

	if results.Entries[0].Feed.FeedURL != testConfig.testFeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, results.Entries[0].Feed.FeedURL)
	}

	if results.Entries[0].Title == "" {
		t.Fatalf(`Invalid title, got empty`)
	}
}

func TestGetAllEntriesEndpointWithFilter(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	feedEntries, err := regularUserClient.Entries(&miniflux.Filter{FeedID: feedID})
	if err != nil {
		t.Fatal(err)
	}

	if len(feedEntries.Entries) == 0 {
		t.Fatalf(`Invalid number of entries, got %d`, len(feedEntries.Entries))
	}

	if feedEntries.Total == 0 {
		t.Fatalf(`Invalid total, got %d`, feedEntries.Total)
	}

	if feedEntries.Entries[0].FeedID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, feedEntries.Entries[0].FeedID)
	}

	if feedEntries.Entries[0].Feed.FeedURL != testConfig.testFeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, feedEntries.Entries[0].Feed.FeedURL)
	}

	if feedEntries.Entries[0].Title == "" {
		t.Fatalf(`Invalid title, got empty`)
	}

	recentEntries, err := regularUserClient.Entries(&miniflux.Filter{Order: "published_at", Direction: "desc"})
	if err != nil {
		t.Fatal(err)
	}

	if len(recentEntries.Entries) == 0 {
		t.Fatalf(`Invalid number of entries, got %d`, len(recentEntries.Entries))
	}

	if recentEntries.Total == 0 {
		t.Fatalf(`Invalid total, got %d`, recentEntries.Total)
	}

	if feedEntries.Entries[0].Title == recentEntries.Entries[0].Title {
		t.Fatalf(`Invalid order, got the same title`)
	}

	searchedEntries, err := regularUserClient.Entries(&miniflux.Filter{Search: "2.0.8"})
	if err != nil {
		t.Fatal(err)
	}

	if searchedEntries.Total != 1 {
		t.Fatalf(`Invalid total, got %d`, searchedEntries.Total)
	}

	if _, err := regularUserClient.Entries(&miniflux.Filter{Status: "invalid"}); err == nil {
		t.Fatal(`Using invalid status should raise an error`)
	}

	if _, err = regularUserClient.Entries(&miniflux.Filter{Direction: "invalid"}); err == nil {
		t.Fatal(`Using invalid direction should raise an error`)
	}

	if _, err = regularUserClient.Entries(&miniflux.Filter{Order: "invalid"}); err == nil {
		t.Fatal(`Using invalid order should raise an error`)
	}
}

func TestGetEntryEndpoints(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := regularUserClient.FeedEntries(feedID, nil)
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}

	entry, err := regularUserClient.FeedEntry(feedID, result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.ID != result.Entries[0].ID {
		t.Fatalf(`Invalid entryID, got %d`, entry.ID)
	}

	if entry.FeedID != feedID {
		t.Fatalf(`Invalid feedID, got %d`, entry.FeedID)
	}

	if entry.Feed.FeedURL != testConfig.testFeedURL {
		t.Fatalf(`Invalid feed URL, got %q`, entry.Feed.FeedURL)
	}

	entry, err = regularUserClient.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.ID != result.Entries[0].ID {
		t.Fatalf(`Invalid entryID, got %d`, entry.ID)
	}

	entry, err = regularUserClient.CategoryEntry(result.Entries[0].Feed.Category.ID, result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.ID != result.Entries[0].ID {
		t.Fatalf(`Invalid entryID, got %d`, entry.ID)
	}
}

func TestUpdateEntryStatusEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := regularUserClient.FeedEntries(feedID, nil)
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}

	if err := regularUserClient.UpdateEntries([]int64{result.Entries[0].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}

	entry, err := regularUserClient.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Status != miniflux.EntryStatusRead {
		t.Fatalf(`Invalid status, got %q`, entry.Status)
	}
}

func TestUpdateEntryEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := regularUserClient.FeedEntries(feedID, nil)
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}

	entryUpdateRequest := &miniflux.EntryModificationRequest{
		Title:   miniflux.SetOptionalField("New title"),
		Content: miniflux.SetOptionalField("New content"),
	}

	updatedEntry, err := regularUserClient.UpdateEntry(result.Entries[0].ID, entryUpdateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if updatedEntry.Title != "New title" {
		t.Errorf(`Invalid title, got %q`, updatedEntry.Title)
	}

	if updatedEntry.Content != "New content" {
		t.Errorf(`Invalid content, got %q`, updatedEntry.Content)
	}

	entry, err := regularUserClient.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Title != "New title" {
		t.Errorf(`Invalid title, got %q`, entry.Title)
	}

	if entry.Content != "New content" {
		t.Errorf(`Invalid content, got %q`, entry.Content)
	}
}

func TestToggleBookmarkEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := regularUserClient.FeedEntries(feedID, &miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}

	if err := regularUserClient.ToggleBookmark(result.Entries[0].ID); err != nil {
		t.Fatal(err)
	}

	entry, err := regularUserClient.Entry(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if !entry.Starred {
		t.Fatalf(`The entry should be bookmarked`)
	}
}

func TestSaveEntryEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := regularUserClient.FeedEntries(feedID, &miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}

	if err := regularUserClient.SaveEntry(result.Entries[0].ID); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Saving an entry should raise a bad request error because no integration is configured`)
	}
}

func TestFetchContentEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := regularUserClient.FeedEntries(feedID, &miniflux.Filter{Limit: 1})
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}

	content, err := regularUserClient.FetchEntryOriginalContent(result.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if content == "" {
		t.Fatalf(`Invalid content, got empty`)
	}
}

func TestFlushHistoryEndpoint(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)

	regularTestUser, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(regularTestUser.ID)

	regularUserClient := miniflux.NewClient(testConfig.testBaseURL, regularTestUser.Username, testConfig.testRegularPassword)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: testConfig.testFeedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := regularUserClient.FeedEntries(feedID, &miniflux.Filter{Limit: 3})
	if err != nil {
		t.Fatalf(`Failed to get entries: %v`, err)
	}

	if err := regularUserClient.UpdateEntries([]int64{result.Entries[0].ID, result.Entries[1].ID}, miniflux.EntryStatusRead); err != nil {
		t.Fatal(err)
	}

	if err := regularUserClient.FlushHistory(); err != nil {
		t.Fatal(err)
	}

	readEntries, err := regularUserClient.Entries(&miniflux.Filter{Status: miniflux.EntryStatusRead})
	if err != nil {
		t.Fatal(err)
	}

	if readEntries.Total != 0 {
		t.Fatalf(`Invalid total, got %d`, readEntries.Total)
	}

	removedEntries, err := regularUserClient.Entries(&miniflux.Filter{Status: miniflux.EntryStatusRemoved})
	if err != nil {
		t.Fatal(err)
	}

	if removedEntries.Total != 2 {
		t.Fatalf(`Invalid total, got %d`, removedEntries.Total)
	}
}
