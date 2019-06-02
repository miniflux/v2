// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config // import "miniflux.app/config"

import (
	"os"
	"testing"
)

func TestGetBooleanValueWithUnsetVariable(t *testing.T) {
	os.Clearenv()
	if getBooleanValue("MY_TEST_VARIABLE") {
		t.Errorf(`Unset variables should returns false`)
	}
}

func TestGetBooleanValue(t *testing.T) {
	scenarios := map[string]bool{
		"":        false,
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
		os.Clearenv()
		os.Setenv("MY_TEST_VARIABLE", input)
		result := getBooleanValue("MY_TEST_VARIABLE")
		if result != expected {
			t.Errorf(`Unexpected result for %q, got %v instead of %v`, input, result, expected)
		}
	}
}

func TestGetStringValueWithUnsetVariable(t *testing.T) {
	os.Clearenv()
	if getStringValue("MY_TEST_VARIABLE", "defaultValue") != "defaultValue" {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestGetStringValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("MY_TEST_VARIABLE", "test")
	if getStringValue("MY_TEST_VARIABLE", "defaultValue") != "test" {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}

func TestGetIntValueWithUnsetVariable(t *testing.T) {
	os.Clearenv()
	if getIntValue("MY_TEST_VARIABLE", 42) != 42 {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestGetIntValueWithInvalidInput(t *testing.T) {
	os.Clearenv()
	os.Setenv("MY_TEST_VARIABLE", "invalid integer")
	if getIntValue("MY_TEST_VARIABLE", 42) != 42 {
		t.Errorf(`Invalid integer should returns the default value`)
	}
}

func TestGetIntValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("MY_TEST_VARIABLE", "2018")
	if getIntValue("MY_TEST_VARIABLE", 42) != 2018 {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}
