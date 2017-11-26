// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// +build integration

package main

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/miniflux/miniflux-go"
)

const (
	testBaseURL          = "http://127.0.0.1:8080"
	testAdminUsername    = "admin"
	testAdminPassword    = "test123"
	testStandardPassword = "secret"
)

func TestWithBadEndpoint(t *testing.T) {
	client := miniflux.NewClient("bad url", testAdminUsername, testAdminPassword)
	_, err := client.Users()
	if err == nil {
		t.Fatal(`Using a bad url should raise an error`)
	}
}

func TestWithWrongCredentials(t *testing.T) {
	client := miniflux.NewClient(testBaseURL, "invalid", "invalid")
	_, err := client.Users()
	if err == nil {
		t.Fatal(`Using bad credentials should raise an error`)
	}
}

func TestGetUsers(t *testing.T) {
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
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
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
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
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.DeleteUser(user.ID); err != nil {
		t.Fatalf(`Unable to remove user: "%v"`, err)
	}
}

func TestGetUser(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	user, err = client.User(user.ID)
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

func TestUpdateUser(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	theme := "black"
	user.Theme = theme
	user, err = client.UpdateUser(user)
	if err != nil {
		t.Fatal(err)
	}

	if user.Theme != theme {
		t.Fatalf(`Unable to update user: got "%v" instead of "%v"`, user.Theme, theme)
	}
}

func TestUpdateUserWithInvalidValue(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	theme := "something that doesn't exists"
	user.Theme = theme
	_, err = client.UpdateUser(user)
	if err == nil {
		t.Fatal(`Updating a user with an invalid value should raise an error`)
	}
}

func TestCannotCreateDuplicateUser(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
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
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	_, err = client.Users()
	if err == nil {
		t.Fatal(`Standard users should not be able to list any users`)
	}
}

func TestCannotGetUserAsNonAdmin(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	_, err = client.User(user.ID)
	if err == nil {
		t.Fatal(`Standard users should not be able to get any users`)
	}
}

func TestCannotUpdateUserAsNonAdmin(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	_, err = client.UpdateUser(user)
	if err == nil {
		t.Fatal(`Standard users should not be able to update any users`)
	}
}

func TestCannotCreateUserAsNonAdmin(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	_, err = client.CreateUser(username, testStandardPassword, false)
	if err == nil {
		t.Fatal(`Standard users should not be able to create users`)
	}
}

func TestCannotDeleteUserAsNonAdmin(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	if err := client.DeleteUser(user.ID); err == nil {
		t.Fatal(`Standard users should not be able to remove any users`)
	}
}

func TestCreateCategory(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	categoryName := "My category"
	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	category, err := client.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	if category.ID == 0 {
		t.Fatalf(`Invalid categoryID, got "%v"`, category.ID)
	}

	if category.UserID != user.ID {
		t.Fatalf(`Invalid userID, got "%v" instead of "%v"`, category.UserID, user.ID)
	}

	if category.Title != categoryName {
		t.Fatalf(`Invalid title, got "%v" instead of "%v"`, category.Title, categoryName)
	}
}

func TestCreateCategoryWithEmptyTitle(t *testing.T) {
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateCategory("")
	if err == nil {
		t.Fatal(`The category title should be mandatory`)
	}
}

func TestCannotCreateDuplicatedCategory(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	categoryName := "My category"
	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	_, err = client.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateCategory(categoryName)
	if err == nil {
		t.Fatal(`Duplicated categories should not be allowed`)
	}
}

func TestUpdateCategory(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	categoryName := "My category"
	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	category, err := client.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	categoryName = "Updated category"
	category, err = client.UpdateCategory(category.ID, categoryName)
	if err != nil {
		t.Fatal(err)
	}

	if category.ID == 0 {
		t.Fatalf(`Invalid categoryID, got "%v"`, category.ID)
	}

	if category.UserID != user.ID {
		t.Fatalf(`Invalid userID, got "%v" instead of "%v"`, category.UserID, user.ID)
	}

	if category.Title != categoryName {
		t.Fatalf(`Invalid title, got "%v" instead of "%v"`, category.Title, categoryName)
	}
}

func TestListCategories(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	user, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	categoryName := "My category"
	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	_, err = client.CreateCategory(categoryName)
	if err != nil {
		t.Fatal(err)
	}

	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	if len(categories) != 2 {
		t.Fatalf(`Invalid number of categories, got "%v" instead of "%v"`, len(categories), 2)
	}

	if categories[0].ID == 0 {
		t.Fatalf(`Invalid categoryID, got "%v"`, categories[0].ID)
	}

	if categories[0].UserID != user.ID {
		t.Fatalf(`Invalid userID, got "%v" instead of "%v"`, categories[0].UserID, user.ID)
	}

	if categories[0].Title != "All" {
		t.Fatalf(`Invalid title, got "%v" instead of "%v"`, categories[0].Title, "All")
	}

	if categories[1].ID == 0 {
		t.Fatalf(`Invalid categoryID, got "%v"`, categories[0].ID)
	}

	if categories[1].UserID != user.ID {
		t.Fatalf(`Invalid userID, got "%v" instead of "%v"`, categories[1].UserID, user.ID)
	}

	if categories[1].Title != categoryName {
		t.Fatalf(`Invalid title, got "%v" instead of "%v"`, categories[1].Title, categoryName)
	}
}

func TestDeleteCategory(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	category, err := client.CreateCategory("My category")
	if err != nil {
		t.Fatal(err)
	}

	err = client.DeleteCategory(category.ID)
	if err != nil {
		t.Fatal(`Removing a category should not raise any error`)
	}
}

func TestCannotDeleteCategoryOfAnotherUser(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	err = client.DeleteCategory(categories[0].ID)
	if err == nil {
		t.Fatal(`Removing a category that belongs to another user should be forbidden`)
	}
}

func TestDiscoverSubscriptions(t *testing.T) {
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	subscriptions, err := client.Discover("https://miniflux.net")
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 1 {
		t.Fatalf(`Invalid number of subscriptions, got "%v" instead of "%v"`, len(subscriptions), 2)
	}

	if subscriptions[0].Title != "Feed" {
		t.Fatalf(`Invalid feed title, got "%v" instead of "%v"`, subscriptions[0].Title, "Feed")
	}

	if subscriptions[0].Type != "atom" {
		t.Fatalf(`Invalid feed type, got "%v" instead of "%v"`, subscriptions[0].Type, "atom")
	}

	if subscriptions[0].URL != "https://miniflux.net/feed" {
		t.Fatalf(`Invalid feed URL, got "%v" instead of "%v"`, subscriptions[0].URL, "https://miniflux.net/feed")
	}
}

func TestCreateFeed(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed("https://miniflux.net/feed", categories[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got "%v"`, feedID)
	}
}

func TestCannotCreateDuplicatedFeed(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	categories, err := client.Categories()
	if err != nil {
		t.Fatal(err)
	}

	feedID, err := client.CreateFeed("https://miniflux.net/feed", categories[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if feedID == 0 {
		t.Fatalf(`Invalid feed ID, got "%v"`, feedID)
	}

	_, err = client.CreateFeed("https://miniflux.net/feed", categories[0].ID)
	if err == nil {
		t.Fatal(`Duplicated feeds should not be allowed`)
	}
}

func TestCreateFeedWithInexistingCategory(t *testing.T) {
	username := getRandomUsername()
	client := miniflux.NewClient(testBaseURL, testAdminUsername, testAdminPassword)
	_, err := client.CreateUser(username, testStandardPassword, false)
	if err != nil {
		t.Fatal(err)
	}

	client = miniflux.NewClient(testBaseURL, username, testStandardPassword)
	_, err = client.CreateFeed("https://miniflux.net/feed", -1)
	if err == nil {
		t.Fatal(`Feeds should not be created with inexisting category`)
	}
}

func getRandomUsername() string {
	rand.Seed(time.Now().UnixNano())
	var suffix []string
	for i := 0; i < 10; i++ {
		suffix = append(suffix, strconv.Itoa(rand.Intn(1000)))
	}
	return "user" + strings.Join(suffix, "")
}
