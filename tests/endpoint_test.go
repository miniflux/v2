// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

//go:build integration
// +build integration

package tests

import (
	"testing"

	miniflux "miniflux.app/client"
)

func TestWithBadEndpoint(t *testing.T) {
	client := miniflux.New("bad url", testAdminUsername, testAdminPassword)
	_, err := client.Users()
	if err == nil {
		t.Fatal(`Using a bad URL should raise an error`)
	}
}
