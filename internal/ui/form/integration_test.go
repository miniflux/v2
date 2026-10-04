// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
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
