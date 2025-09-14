// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"strings"
	"testing"
)

func TestValidateChoices(t *testing.T) {
	tests := []struct {
		name        string
		rawValue    string
		choices     []string
		expectError bool
	}{
		{
			name:        "valid choice",
			rawValue:    "option1",
			choices:     []string{"option1", "option2", "option3"},
			expectError: false,
		},
		{
			name:        "valid choice from middle",
			rawValue:    "option2",
			choices:     []string{"option1", "option2", "option3"},
			expectError: false,
		},
		{
			name:        "valid choice from end",
			rawValue:    "option3",
			choices:     []string{"option1", "option2", "option3"},
			expectError: false,
		},
		{
			name:        "invalid choice",
			rawValue:    "invalid",
			choices:     []string{"option1", "option2", "option3"},
			expectError: true,
		},
		{
			name:        "empty value with non-empty choices",
			rawValue:    "",
			choices:     []string{"option1", "option2"},
			expectError: true,
		},
		{
			name:        "case sensitive - different case",
			rawValue:    "OPTION1",
			choices:     []string{"option1", "option2"},
			expectError: true,
		},
		{
			name:        "single choice valid",
			rawValue:    "only",
			choices:     []string{"only"},
			expectError: false,
		},
		{
			name:        "empty choices list",
			rawValue:    "anything",
			choices:     []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChoices(tt.rawValue, tt.choices)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else {
					// Verify error message format
					expectedPrefix := "value must be one of:"
					if !strings.Contains(err.Error(), expectedPrefix) {
						t.Errorf("error message should contain '%s', got: %s", expectedPrefix, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateListChoices(t *testing.T) {
	tests := []struct {
		name        string
		inputValues []string
		choices     []string
		expectError bool
	}{
		{
			name:        "all valid choices",
			inputValues: []string{"option1", "option2"},
			choices:     []string{"option1", "option2", "option3"},
			expectError: false,
		},
		{
			name:        "single valid choice",
			inputValues: []string{"option1"},
			choices:     []string{"option1", "option2", "option3"},
			expectError: false,
		},
		{
			name:        "empty input list",
			inputValues: []string{},
			choices:     []string{"option1", "option2", "option3"},
			expectError: false,
		},
		{
			name:        "all choices from available list",
			inputValues: []string{"option1", "option2", "option3"},
			choices:     []string{"option1", "option2", "option3"},
			expectError: false,
		},
		{
			name:        "duplicate valid choices",
			inputValues: []string{"option1", "option1", "option2"},
			choices:     []string{"option1", "option2", "option3"},
			expectError: false,
		},
		{
			name:        "one invalid choice",
			inputValues: []string{"option1", "invalid"},
			choices:     []string{"option1", "option2", "option3"},
			expectError: true,
		},
		{
			name:        "all invalid choices",
			inputValues: []string{"invalid1", "invalid2"},
			choices:     []string{"option1", "option2", "option3"},
			expectError: true,
		},
		{
			name:        "case sensitive - different case",
			inputValues: []string{"OPTION1"},
			choices:     []string{"option1", "option2"},
			expectError: true,
		},
		{
			name:        "empty string in input",
			inputValues: []string{""},
			choices:     []string{"option1", "option2"},
			expectError: true,
		},
		{
			name:        "empty choices list with non-empty input",
			inputValues: []string{"anything"},
			choices:     []string{},
			expectError: true,
		},
		{
			name:        "mixed valid and invalid choices",
			inputValues: []string{"option1", "invalid", "option2"},
			choices:     []string{"option1", "option2", "option3"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateListChoices(tt.inputValues, tt.choices)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else {
					// Verify error message format
					expectedPrefix := "value must be one of:"
					if !strings.Contains(err.Error(), expectedPrefix) {
						t.Errorf("error message should contain '%s', got: %s", expectedPrefix, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateGreaterThan(t *testing.T) {
	if err := validateGreaterThan("10", 5); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if err := validateGreaterThan("5", 5); err == nil {
		t.Errorf("expected error, got none")
	}

	if err := validateGreaterThan("abc", 5); err == nil {
		t.Errorf("expected error for non-integer input, got none")
	}

	if err := validateGreaterThan("-1", 0); err == nil {
		t.Errorf("expected error for value below minimum, got none")
	}
}

func TestValidateGreaterOrEqualThan(t *testing.T) {
	if err := validateGreaterOrEqualThan("10", 5); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if err := validateGreaterOrEqualThan("5", 5); err != nil {
		t.Errorf("expected no error for equal value, got: %v", err)
	}

	if err := validateGreaterOrEqualThan("abc", 5); err == nil {
		t.Errorf("expected error for non-integer input, got none")
	}

	if err := validateGreaterOrEqualThan("-1", 0); err == nil {
		t.Errorf("expected error for value below minimum, got none")
	}
}

func TestValidateRange(t *testing.T) {
	tests := []struct {
		name        string
		rawValue    string
		min         int
		max         int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid integer within range",
			rawValue:    "5",
			min:         1,
			max:         10,
			expectError: false,
		},
		{
			name:        "valid integer at minimum",
			rawValue:    "1",
			min:         1,
			max:         10,
			expectError: false,
		},
		{
			name:        "valid integer at maximum",
			rawValue:    "10",
			min:         1,
			max:         10,
			expectError: false,
		},
		{
			name:        "valid zero in range",
			rawValue:    "0",
			min:         -5,
			max:         5,
			expectError: false,
		},
		{
			name:        "valid negative in range",
			rawValue:    "-3",
			min:         -5,
			max:         5,
			expectError: false,
		},
		{
			name:        "integer below minimum",
			rawValue:    "0",
			min:         1,
			max:         10,
			expectError: true,
			errorMsg:    "value must be between 1 and 10",
		},
		{
			name:        "integer above maximum",
			rawValue:    "11",
			min:         1,
			max:         10,
			expectError: true,
			errorMsg:    "value must be between 1 and 10",
		},
		{
			name:        "integer far below minimum",
			rawValue:    "-100",
			min:         1,
			max:         10,
			expectError: true,
			errorMsg:    "value must be between 1 and 10",
		},
		{
			name:        "integer far above maximum",
			rawValue:    "100",
			min:         1,
			max:         10,
			expectError: true,
			errorMsg:    "value must be between 1 and 10",
		},
		{
			name:        "non-integer string",
			rawValue:    "abc",
			min:         1,
			max:         10,
			expectError: true,
			errorMsg:    "value must be an integer",
		},
		{
			name:        "empty string",
			rawValue:    "",
			min:         1,
			max:         10,
			expectError: true,
			errorMsg:    "value must be an integer",
		},
		{
			name:        "float string",
			rawValue:    "5.5",
			min:         1,
			max:         10,
			expectError: true,
			errorMsg:    "value must be an integer",
		},
		{
			name:        "string with spaces",
			rawValue:    " 5 ",
			min:         1,
			max:         10,
			expectError: true,
			errorMsg:    "value must be an integer",
		},
		{
			name:        "single value range",
			rawValue:    "5",
			min:         5,
			max:         5,
			expectError: false,
		},
		{
			name:        "single value range - below",
			rawValue:    "4",
			min:         5,
			max:         5,
			expectError: true,
			errorMsg:    "value must be between 5 and 5",
		},
		{
			name:        "single value range - above",
			rawValue:    "6",
			min:         5,
			max:         5,
			expectError: true,
			errorMsg:    "value must be between 5 and 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRange(tt.rawValue, tt.min, tt.max)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}
