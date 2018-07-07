// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"testing"
)

func TestDebugModeOn(t *testing.T) {
	os.Clearenv()
	os.Setenv("DEBUG", "1")
	cfg := NewConfig()

	if !cfg.HasDebugMode() {
		t.Fatalf(`Unexpected debug mode value, got "%v"`, cfg.HasDebugMode())
	}
}

func TestDebugModeOff(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()

	if cfg.HasDebugMode() {
		t.Fatalf(`Unexpected debug mode value, got "%v"`, cfg.HasDebugMode())
	}
}

func TestCustomBaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example.org")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://example.org" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != "http://example.org" {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestCustomBaseURLWithTrailingSlash(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example.org/folder/")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://example.org/folder" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != "http://example.org" {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "/folder" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestBaseURLWithoutScheme(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "example.org/folder/")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://localhost" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != "http://localhost" {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestBaseURLWithInvalidScheme(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "ftp://example.org/folder/")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://localhost" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != "http://localhost" {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestDefaultBaseURL(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()

	if cfg.BaseURL() != "http://localhost" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != "http://localhost" {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestHSTSOn(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()

	if !cfg.HasHSTS() {
		t.Fatalf(`Unexpected HSTS value, got "%v"`, cfg.HasHSTS())
	}
}

func TestHSTSOff(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISABLE_HSTS", "1")
	cfg := NewConfig()

	if cfg.HasHSTS() {
		t.Fatalf(`Unexpected HSTS value, got "%v"`, cfg.HasHSTS())
	}
}
