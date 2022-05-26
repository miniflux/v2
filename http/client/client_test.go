// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package client // import "miniflux.app/http/client"

import (
	"testing"
)

func TestClientWithDelay(t *testing.T) {
	clt := New("http://httpbin.org/delay/5")
	clt.ClientTimeout = 1
	_, err := clt.Get()
	if err == nil {
		t.Fatal(`The client should stops after 1 second`)
	}
}

func TestClientWithError(t *testing.T) {
	clt := New("http://httpbin.org/status/502")
	clt.ClientTimeout = 5
	response, err := clt.Get()
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != 502 {
		t.Fatalf(`Unexpected response status code: %d`, response.StatusCode)
	}

	if !response.HasServerFailure() {
		t.Fatal(`A 502 error is considered as server failure`)
	}
}

func TestClientWithResponseTooLarge(t *testing.T) {
	clt := New("http://httpbin.org/bytes/100")
	clt.ClientMaxBodySize = 10
	_, err := clt.Get()
	if err == nil {
		t.Fatal(`The client should fails when reading a response too large`)
	}
}

func TestClientWithBasicAuth(t *testing.T) {
	clt := New("http://httpbin.org/basic-auth/testuser/testpassword")
	clt.WithCredentials("testuser", "testpassword")
	_, err := clt.Get()
	if err != nil {
		t.Fatalf(`The client should be authenticated successfully: %v`, err)
	}
}
