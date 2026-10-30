// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestAuthSettingsValid(t *testing.T) {
	form := &AuthSettingsForm{
		Username:     "user",
		Password:     "hunter2",
		Confirmation: "hunter2",
	}

	if err := form.Validate(); err != nil {
		t.Error(err)
	}
}

func TestAuthSettingsUsernameEmpty(t *testing.T) {
	form := &AuthSettingsForm{
		Username: "",
	}

	if err := form.Validate(); err == nil {
		t.Error("Validate should return an error")
	}
}

func TestAuthSettingsConfirmationEmpty(t *testing.T) {
	form := &AuthSettingsForm{
		Username:     "user",
		Password:     "hunter2",
		Confirmation: "",
	}

	if err := form.Validate(); err != nil {
		t.Error(err)
	}

	if form.Password != "" {
		t.Error("Password should have been cleared")
	}
}

func TestAuthSettingsConfirmationIncorrect(t *testing.T) {
	form := &AuthSettingsForm{
		Username:     "user",
		Password:     "hunter2",
		Confirmation: "unter2",
	}

	if err := form.Validate(); err == nil {
		t.Error("Validate should return an error")
	}
}

func TestAuthSettingsMergeKeepsPasswordWhenEmpty(t *testing.T) {
	user := &model.User{Username: "old", Password: "existing"}
	form := &AuthSettingsForm{Username: "new"}

	form.Merge(user)

	if user.Username != "new" {
		t.Errorf("Expected username to be updated, got %q", user.Username)
	}

	if user.Password != "existing" {
		t.Errorf("Expected password to be unchanged, got %q", user.Password)
	}
}

func TestNewAuthSettingsFormReadsCurrentPassword(t *testing.T) {
	values := url.Values{
		"username":         {"user"},
		"current_password": {"old-secret"},
		"password":         {"new-secret"},
		"confirmation":     {"new-secret"},
	}
	request := httptest.NewRequest(http.MethodPost, "/authentication", strings.NewReader(values.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	form := NewAuthSettingsForm(request)

	if form.CurrentPassword != "old-secret" {
		t.Errorf("Expected current password to be read from the request, got %q", form.CurrentPassword)
	}
}
