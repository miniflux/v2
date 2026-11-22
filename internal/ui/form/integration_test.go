// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestValidateGoogleReader(t *testing.T) {
	scenarios := []struct {
		name               string
		form               IntegrationForm
		storedPasswordHash string
		expectError        bool
	}{
		{
			name: "disabled without credentials",
			form: IntegrationForm{},
		},
		{
			name: "enabled with username and new password",
			form: IntegrationForm{GoogleReaderEnabled: true, GoogleReaderUsername: "user", GoogleReaderPassword: "secret"},
		},
		{
			name:               "enabled with username and stored password",
			form:               IntegrationForm{GoogleReaderEnabled: true, GoogleReaderUsername: "user"},
			storedPasswordHash: "$2a$10$hash",
		},
		{
			name:        "enabled without password",
			form:        IntegrationForm{GoogleReaderEnabled: true, GoogleReaderUsername: "user"},
			expectError: true,
		},
		{
			name:               "enabled without username",
			form:               IntegrationForm{GoogleReaderEnabled: true, GoogleReaderPassword: "secret"},
			storedPasswordHash: "$2a$10$hash",
			expectError:        true,
		},
		{
			name:        "enabled without username and password",
			form:        IntegrationForm{GoogleReaderEnabled: true},
			expectError: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			validationErr := scenario.form.ValidateGoogleReader(scenario.storedPasswordHash)
			if scenario.expectError && validationErr == nil {
				t.Fatal("expected a validation error")
			}
			if !scenario.expectError && validationErr != nil {
				t.Fatalf("unexpected validation error: %v", validationErr)
			}
		})
	}
}

func TestNewIntegrationFormLinktacoVisibility(t *testing.T) {
	scenarios := []struct {
		name     string
		values   url.Values
		expected string
	}{
		{name: "missing field", values: url.Values{}, expected: "PUBLIC"},
		{name: "empty value", values: url.Values{"linktaco_visibility": {""}}, expected: "PUBLIC"},
		{name: "public", values: url.Values{"linktaco_visibility": {"PUBLIC"}}, expected: "PUBLIC"},
		{name: "private", values: url.Values{"linktaco_visibility": {"PRIVATE"}}, expected: "PRIVATE"},
		{name: "invalid value", values: url.Values{"linktaco_visibility": {"UNLISTED"}}, expected: "PUBLIC"},
		{name: "lowercase value", values: url.Values{"linktaco_visibility": {"private"}}, expected: "PUBLIC"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/integration", strings.NewReader(scenario.values.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			integrationForm := NewIntegrationForm(request)
			if integrationForm.LinktacoVisibility != scenario.expected {
				t.Fatalf("expected visibility %q, got %q", scenario.expected, integrationForm.LinktacoVisibility)
			}
		})
	}
}
