// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"slices"
	"testing"
)

func TestGroupsFromClaims(t *testing.T) {
	scenarios := []struct {
		name           string
		claims         map[string]any
		expectedGroups []string
		expectedFound  bool
		expectedError  bool
	}{
		{
			name:   "missing claim",
			claims: map[string]any{"email": "user@example.org"},
		},
		{
			name:           "single group",
			claims:         map[string]any{"groups": "admins"},
			expectedGroups: []string{"admins"},
			expectedFound:  true,
		},
		{
			name:           "list of groups",
			claims:         map[string]any{"groups": []any{"users", "admins"}},
			expectedGroups: []string{"users", "admins"},
			expectedFound:  true,
		},
		{
			name:           "empty list",
			claims:         map[string]any{"groups": []any{}},
			expectedGroups: []string{},
			expectedFound:  true,
		},
		{
			name:          "list with a non-string entry",
			claims:        map[string]any{"groups": []any{"admins", 42.0}},
			expectedFound: true,
			expectedError: true,
		},
		{
			name:          "unsupported type",
			claims:        map[string]any{"groups": map[string]any{"name": "admins"}},
			expectedFound: true,
			expectedError: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			groups, found, err := groupsFromClaims(scenario.claims, "groups")

			if (err != nil) != scenario.expectedError {
				t.Fatalf("Expected error %v, got %v", scenario.expectedError, err)
			}

			if found != scenario.expectedFound {
				t.Errorf("Expected found %v, got %v", scenario.expectedFound, found)
			}

			if !slices.Equal(groups, scenario.expectedGroups) {
				t.Errorf("Expected groups %v, got %v", scenario.expectedGroups, groups)
			}
		})
	}
}
