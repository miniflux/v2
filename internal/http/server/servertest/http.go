// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package servertest // import "miniflux.app/v2/internal/http/server/servertest"

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

// httpClient fails a request after a timeout, so that a stuck handler fails one test instead of the whole run.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// Send sends the request and returns the status and body of the response.
func Send(t testing.TB, request *http.Request) (int, []byte) {
	t.Helper()

	response, err := httpClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response.StatusCode, body
}

// DecodeJSON decodes the body of a successful response into result, and fails the test when the status is not 200.
func DecodeJSON(t testing.TB, status int, body []byte, result any) {
	t.Helper()

	if status != http.StatusOK {
		t.Fatalf(`Expected status 200, got %d: %s`, status, body)
	}
	if err := json.Unmarshal(body, result); err != nil {
		t.Fatalf(`Unable to decode %s: %v`, body, err)
	}
}
