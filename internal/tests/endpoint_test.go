// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package tests

import (
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestWithBadEndpoint(t *testing.T) {
	client := miniflux.New("bad url", testAdminUsername, testAdminPassword)
	_, err := client.Users()
	if err == nil {
		t.Fatal(`Using a bad URL should raise an error`)
	}
}
