// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client // import "miniflux.app/http/client"

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mccutchen/go-httpbin/v2/httpbin"
)

var srv *httptest.Server

func TestMain(m *testing.M) {
	srv = httptest.NewServer(httpbin.New())
	exitCode := m.Run()
	srv.Close()
	os.Exit(exitCode)
}

func MakeClient(path string) *Client {
	return New(fmt.Sprintf("%s%s", srv.URL, path))
}

func TestClientWithDelay(t *testing.T) {
	clt := MakeClient("/delay/5")
	clt.ClientTimeout = 1
	_, err := clt.Get()
	if err == nil {
		t.Fatal(`The client should stops after 1 second`)
	}
}

func TestClientWithError(t *testing.T) {
	clt := MakeClient("/status/502")
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
	clt := MakeClient("/bytes/100")
	clt.ClientMaxBodySize = 10
	_, err := clt.Get()
	if err == nil {
		t.Fatal(`The client should fails when reading a response too large`)
	}
}

func TestClientWithBasicAuth(t *testing.T) {
	clt := MakeClient("/basic-auth/testuser/testpassword")
	clt.WithCredentials("testuser", "testpassword")
	_, err := clt.Get()
	if err != nil {
		t.Fatalf(`The client should be authenticated successfully: %v`, err)
	}
}
