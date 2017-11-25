// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// +build integration

package main

import (
	"testing"

	"github.com/miniflux/miniflux-go"
)

const (
	testBaseURL  = "http://127.0.0.1:8080"
	testUsername = "admin"
	testPassword = "test123"
)

func TestGetUsers(t *testing.T) {
	client := miniflux.NewClient(testBaseURL, testUsername, testPassword)
	users, err := client.Users()
	if err != nil {
		t.Fatal(err)
		return
	}

	if len(users) == 0 {
		t.Fatal("The list of users is empty")
	}

	if users[0].ID == 0 {
		t.Fatalf(`Invalid userID, got %v`, users[0].ID)
	}

	if users[0].Username != testUsername {
		t.Fatalf(`Invalid username, got %v`, users[0].Username)
	}

	if users[0].Password != "" {
		t.Fatalf(`Invalid password, got %v`, users[0].Password)
	}

	if users[0].Language != "en_US" {
		t.Fatalf(`Invalid language, got %v`, users[0].Language)
	}

	if users[0].Theme != "default" {
		t.Fatalf(`Invalid theme, got %v`, users[0].Theme)
	}

	if users[0].Timezone != "UTC" {
		t.Fatalf(`Invalid timezone, got %v`, users[0].Timezone)
	}

	if !users[0].IsAdmin {
		t.Fatalf(`Invalid role, got %v`, users[0].IsAdmin)
	}
}

func TestCreateStandardUser(t *testing.T) {
	client := miniflux.NewClient(testBaseURL, testUsername, testPassword)
	user, err := client.CreateUser("test", "test123", false)
	if err != nil {
		t.Fatal(err)
		return
	}

	if user.ID == 0 {
		t.Fatalf(`Invalid userID, got %v`, user.ID)
	}

	if user.Username != "test" {
		t.Fatalf(`Invalid username, got %v`, user.Username)
	}

	if user.Password != "" {
		t.Fatalf(`Invalid password, got %v`, user.Password)
	}

	if user.Language != "en_US" {
		t.Fatalf(`Invalid language, got %v`, user.Language)
	}

	if user.Theme != "default" {
		t.Fatalf(`Invalid theme, got %v`, user.Theme)
	}

	if user.Timezone != "UTC" {
		t.Fatalf(`Invalid timezone, got %v`, user.Timezone)
	}

	if user.IsAdmin {
		t.Fatalf(`Invalid role, got %v`, user.IsAdmin)
	}

	if user.LastLoginAt != nil {
		t.Fatalf(`Invalid last login date, got %v`, user.LastLoginAt)
	}
}
