// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package tests

import (
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestVersionEndpoint(t *testing.T) {
	client := miniflux.New(testBaseURL, testAdminUsername, testAdminPassword)
	version, err := client.Version()
	if err != nil {
		t.Fatal(err)
	}

	if version.Version == "" {
		t.Fatal(`Version should not be empty`)
	}

	if version.Commit == "" {
		t.Fatal(`Commit should not be empty`)
	}

	if version.BuildDate == "" {
		t.Fatal(`Build date should not be empty`)
	}

	if version.GoVersion == "" {
		t.Fatal(`Go version should not be empty`)
	}

	if version.Compiler == "" {
		t.Fatal(`Compiler should not be empty`)
	}

	if version.Arch == "" {
		t.Fatal(`Arch should not be empty`)
	}

	if version.OS == "" {
		t.Fatal(`OS should not be empty`)
	}
}
