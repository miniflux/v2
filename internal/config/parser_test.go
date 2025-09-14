// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestParseStringValue(t *testing.T) {
	// Test with non-empty value
	result := parseStringValue("test", "fallback")
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}

	// Test with empty value
	result = parseStringValue("", "fallback")
	if result != "fallback" {
		t.Errorf("Expected 'fallback', got '%s'", result)
	}

	// Test with empty value and empty fallback
	result = parseStringValue("", "")
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestParseBoolValue(t *testing.T) {
	// Test with empty value - should return fallback
	result, err := parseBoolValue("", true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}

	// Test true values
	trueValues := []string{"1", "yes", "true", "on", "YES", "TRUE", "ON"}
	for _, value := range trueValues {
		result, err := parseBoolValue(value, false)
		if err != nil {
			t.Errorf("Unexpected error for value '%s': %v", value, err)
		}
		if result != true {
			t.Errorf("Expected true for '%s', got %v", value, result)
		}
	}

	// Test false values
	falseValues := []string{"0", "no", "false", "off", "NO", "FALSE", "OFF"}
	for _, value := range falseValues {
		result, err := parseBoolValue(value, true)
		if err != nil {
			t.Errorf("Unexpected error for value '%s': %v", value, err)
		}
		if result != false {
			t.Errorf("Expected false for '%s', got %v", value, result)
		}
	}

	// Test invalid value - should return error
	_, err = parseBoolValue("invalid", false)
	if err == nil {
		t.Error("Expected error for invalid boolean value")
	}
}

func TestParseIntValue(t *testing.T) {
	// Test with empty value - should return fallback
	result := parseIntValue("", 42)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test with valid integer
	result = parseIntValue("123", 42)
	if result != 123 {
		t.Errorf("Expected 123, got %d", result)
	}

	// Test with invalid integer - should return fallback
	result = parseIntValue("invalid", 42)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test with zero
	result = parseIntValue("0", 42)
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
}

func TestParsedInt64Value(t *testing.T) {
	// Test with empty value - should return fallback
	result := ParsedInt64Value("", 42)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test with valid int64
	result = ParsedInt64Value("9223372036854775807", 42)
	if result != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807, got %d", result)
	}

	// Test with invalid int64 - should return fallback
	result = ParsedInt64Value("invalid", 42)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestParseStringListValue(t *testing.T) {
	// Test with empty value - should return fallback
	fallback := []string{"a", "b"}
	result := parseStringListValue("", fallback)
	if !reflect.DeepEqual(result, fallback) {
		t.Errorf("Expected %v, got %v", fallback, result)
	}

	// Test with single value
	result = parseStringListValue("item1", nil)
	expected := []string{"item1"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with multiple values
	result = parseStringListValue("item1,item2,item3", nil)
	expected = []string{"item1", "item2", "item3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with duplicates - should remove duplicates
	result = parseStringListValue("item1,item2,item1", nil)
	expected = []string{"item1", "item2"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with spaces
	result = parseStringListValue(" item1 , item2 , item3 ", nil)
	expected = []string{"item1", "item2", "item3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestParseDurationValue(t *testing.T) {
	// Test with empty value - should return fallback
	fallback := 5 * time.Second
	result := parseDurationValue("", time.Second, fallback)
	if result != fallback {
		t.Errorf("Expected %v, got %v", fallback, result)
	}

	// Test with valid duration
	result = parseDurationValue("30", time.Second, fallback)
	expected := 30 * time.Second
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with minutes
	result = parseDurationValue("5", time.Minute, fallback)
	expected = 5 * time.Minute
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with invalid value - should return fallback
	result = parseDurationValue("invalid", time.Second, fallback)
	if result != fallback {
		t.Errorf("Expected %v, got %v", fallback, result)
	}
}

func TestParseURLValue(t *testing.T) {
	// Test with empty value - should return fallback
	fallbackURL, _ := url.Parse("https://fallback.com")
	result, err := parseURLValue("", fallbackURL)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != fallbackURL {
		t.Errorf("Expected %v, got %v", fallbackURL, result)
	}

	// Test with valid URL
	result, err = parseURLValue("https://example.com", nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.String() != "https://example.com" {
		t.Errorf("Expected https://example.com, got %s", result.String())
	}

	// Test with invalid URL - should return fallback and error
	result, err = parseURLValue("://invalid", fallbackURL)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
	if result != fallbackURL {
		t.Errorf("Expected fallback URL, got %v", result)
	}
}

func TestConfigFileParsing(t *testing.T) {
	fileContent := `
		# This is a comment
		LOG_FILE=miniflux.log
		LOG_DATE_TIME=1
		LOG_FORMAT=json
		LISTEN_ADDR=:8080,:8443
	`

	// Write a temporary config file and parse it
	tmpFile, err := os.CreateTemp("", "miniflux-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	filename := tmpFile.Name()
	if _, err := tmpFile.WriteString(fileContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	configParser := NewConfigParser()
	configOptions, err := configParser.ParseFile(filename)
	if err != nil {
		t.Fatalf("Unexpected parsing error: %v", err)
	}

	if configOptions.LogFile() != "miniflux.log" {
		t.Fatalf("Unexpected log file, got %q", configOptions.LogFile())
	}

	if configOptions.LogDateTime() != true {
		t.Fatalf("Unexpected log datetime, got %v", configOptions.LogDateTime())
	}

	if configOptions.LogFormat() != "json" {
		t.Fatalf("Unexpected log format, got %q", configOptions.LogFormat())
	}

	if configOptions.LogLevel() != "info" {
		t.Fatalf("Unexpected log level, got %q", configOptions.LogLevel())
	}

	if len(configOptions.ListenAddr()) != 2 || configOptions.ListenAddr()[0] != ":8080" || configOptions.ListenAddr()[1] != ":8443" {
		t.Fatalf("Unexpected listen addresses, got %v", configOptions.ListenAddr())
	}
}

func TestConfigFileParsingWithIncorrectKeyValuePair(t *testing.T) {
	fileContent := `
		LOG_FILE=miniflux.log
		INVALID_LINE
	`

	// Write a temporary config file and parse it
	tmpFile, err := os.CreateTemp("", "miniflux-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	filename := tmpFile.Name()
	if _, err := tmpFile.WriteString(fileContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	configParser := NewConfigParser()
	_, err = configParser.ParseFile(filename)
	if err != nil {
		t.Fatal("Invalid lines should be ignored, but got error:", err)
	}
}

func TestParseAdminPasswordFileOption(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "password-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	password := "supersecret"
	if _, err := tmpFile.WriteString(password); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	os.Clearenv()
	os.Setenv("ADMIN_PASSWORD_FILE", tmpFile.Name())

	configParser := NewConfigParser()
	configOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unexpected parsing error: %v", err)
	}

	if configOptions.AdminPassword() != password {
		t.Fatalf("Unexpected admin password, got %q", configOptions.AdminPassword())
	}
}

func TestParseAdminPasswordFileOptionWithEmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "empty-password-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	os.Clearenv()
	os.Setenv("ADMIN_PASSWORD_FILE", tmpFile.Name())

	configParser := NewConfigParser()
	_, err = configParser.ParseEnvironmentVariables()
	if err == nil {
		t.Fatal("Expected error due to empty password file, but got none")
	}
}

func TestParseLogFileOptionDefaultValue(t *testing.T) {
	os.Clearenv()

	configParser := NewConfigParser()
	configOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unexpected parsing error: %v", err)
	}

	if configOptions.LogFile() != "stderr" {
		t.Fatalf("Unexpected default log file, got %q", configOptions.LogFile())
	}
}

func TestParseLogFileOptionWithCustomFilename(t *testing.T) {
	os.Clearenv()
	os.Setenv("LOG_FILE", "miniflux.log")

	configParser := NewConfigParser()
	configOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unexpected parsing error: %v", err)
	}

	if configOptions.LogFile() != "miniflux.log" {
		t.Fatalf("Unexpected log file, got %q", configOptions.LogFile())
	}
}

func TestParseLogFileOptionWithEmptyValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("LOG_FILE", "")

	configParser := NewConfigParser()
	configOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unexpected parsing error: %v", err)
	}

	if configOptions.LogFile() != "stderr" {
		t.Fatalf("Unexpected log file, got %q", configOptions.LogFile())
	}
}

func TestParseLogDateTimeOptionDefaultValue(t *testing.T) {
	os.Clearenv()

	configParser := NewConfigParser()
	configOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unexpected parsing error: %v", err)
	}

	if configOptions.LogDateTime() != false {
		t.Fatalf("Unexpected default log datetime, got %v", configOptions.LogDateTime())
	}
}

func TestParseLogDateTimeOptionWithCustomValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("LOG_DATE_TIME", "true")

	configParser := NewConfigParser()
	configOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unexpected parsing error: %v", err)
	}

	if configOptions.LogDateTime() != true {
		t.Fatalf("Unexpected log datetime, got %v", configOptions.LogDateTime())
	}
}

func TestParseLogDateTimeOptionWithEmptyValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("LOG_DATE_TIME", "")

	configParser := NewConfigParser()
	configOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unexpected parsing error: %v", err)
	}

	if configOptions.LogDateTime() != false {
		t.Fatalf("Unexpected log datetime, got %v", configOptions.LogDateTime())
	}
}

func TestParseLogDateTimeOptionWithIncorrectValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("LOG_DATE_TIME", "invalid")

	configParser := NewConfigParser()
	if _, err := configParser.ParseEnvironmentVariables(); err == nil {
		t.Fatal("Expected parsing error, got nil")
	}
}
