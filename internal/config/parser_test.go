// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"testing"
)

func TestParseBoolValue(t *testing.T) {
	scenarios := map[string]bool{
		"":        true,
		"1":       true,
		"Yes":     true,
		"yes":     true,
		"True":    true,
		"true":    true,
		"on":      true,
		"false":   false,
		"off":     false,
		"invalid": false,
	}

	for input, expected := range scenarios {
		result := parseBool(input, true)
		if result != expected {
			t.Errorf(`Unexpected result for %q, got %v instead of %v`, input, result, expected)
		}
	}
}

func TestParseStringValueWithUnsetVariable(t *testing.T) {
	if parseString("", "defaultValue") != "defaultValue" {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestParseStringValue(t *testing.T) {
	if parseString("test", "defaultValue") != "test" {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}

func TestParseIntValueWithUnsetVariable(t *testing.T) {
	if parseInt("", 42) != 42 {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestParseIntValueWithInvalidInput(t *testing.T) {
	if parseInt("invalid integer", 42) != 42 {
		t.Errorf(`Invalid integer should returns the default value`)
	}
}

func TestParseIntValue(t *testing.T) {
	if parseInt("2018", 42) != 2018 {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}
