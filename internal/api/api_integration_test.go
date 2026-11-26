// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestIncorrectEndpoint(t *testing.T) {
	t.Parallel()

	requireIntegrationTestConfig(t)

	client := miniflux.NewClient("incorrect url")
	if _, err := client.Users(); err == nil {
		t.Fatal(`Using an incorrect URL should raise an error`)
	}

	client = miniflux.NewClient("")
	if _, err := client.Users(); err == nil {
		t.Fatal(`Using an empty URL should raise an error`)
	}
}

func TestHealthcheckEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := miniflux.NewClient(testConfig.testBaseURL)
	if err := client.Healthcheck(); err != nil {
		t.Fatal(err)
	}
}

func TestVersionEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := testConfig.newAdminClient()
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

func TestInvalidCredentials(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := miniflux.NewClient(testConfig.testBaseURL, "invalid", "invalid")
	_, err := client.Users()
	if !errors.Is(err, miniflux.ErrNotAuthorized) {
		t.Fatalf(`Using bad credentials should raise a "not authorized" error, got %v`, err)
	}
}
