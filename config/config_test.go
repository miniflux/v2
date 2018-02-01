// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"testing"
)

func TestGetCustomBaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example.org")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://example.org" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}
}

func TestGetCustomBaseURLWithTrailingSlash(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example.org/folder/")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://example.org/folder" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}
}

func TestGetDefaultBaseURL(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()

	if cfg.BaseURL() != "http://localhost" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}
}
