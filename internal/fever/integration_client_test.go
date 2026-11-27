// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// feverAPIVersion is the API version that every Fever response reports.
const feverAPIVersion = 3

// feverClient calls the Fever API. It sends no API key when apiKey is empty.
//
// Like Fever clients, it always calls the endpoint with the api parameter, which has no value.
type feverClient struct {
	baseURL string
	apiKey  string
}

// get sends a GET request with the API key and the parameters in the query string.
func (client *feverClient) get(t *testing.T, query url.Values) (int, []byte) {
	t.Helper()

	return client.getRawQuery(t, client.withAPIKey(query).Encode())
}

// getRawQuery sends a GET request with the given query string, without adding the API key.
// It sends selectors without a value, such as "groups" rather than "groups=", as Fever clients do.
func (client *feverClient) getRawQuery(t *testing.T, rawQuery string) (int, []byte) {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, client.endpoint(rawQuery), nil)
	if err != nil {
		t.Fatal(err)
	}
	return send(t, request)
}

// post sends a form with the API key and the parameters in the request body.
func (client *feverClient) post(t *testing.T, form url.Values) (int, []byte) {
	t.Helper()

	body := strings.NewReader(client.withAPIKey(form).Encode())
	request, err := http.NewRequest(http.MethodPost, client.endpoint(""), body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return send(t, request)
}

// getJSON sends a GET request and decodes the authenticated response into result.
func (client *feverClient) getJSON(t *testing.T, query url.Values, result any) {
	t.Helper()

	status, body := client.get(t, query)
	decodeAuthenticatedJSON(t, status, body, result)
}

// postJSON sends a form and decodes the authenticated response into result.
func (client *feverClient) postJSON(t *testing.T, form url.Values, result any) {
	t.Helper()

	status, body := client.post(t, form)
	decodeAuthenticatedJSON(t, status, body, result)
}

func (client *feverClient) endpoint(rawQuery string) string {
	if rawQuery == "" {
		return client.baseURL + "/fever/?api"
	}
	return client.baseURL + "/fever/?api&" + rawQuery
}

// withAPIKey returns a copy of the parameters with the API key of the client, when it has one.
func (client *feverClient) withAPIKey(params url.Values) url.Values {
	values := url.Values{}
	for key, list := range params {
		values[key] = list
	}
	if client.apiKey != "" {
		values.Set("api_key", client.apiKey)
	}
	return values
}

func send(t *testing.T, request *http.Request) (int, []byte) {
	t.Helper()

	response, err := http.DefaultClient.Do(request)
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

func decodeJSON(t *testing.T, status int, body []byte, result any) {
	t.Helper()

	if status != http.StatusOK {
		t.Fatalf(`Expected status 200, got %d: %s`, status, body)
	}
	if err := json.Unmarshal(body, result); err != nil {
		t.Fatalf(`Unable to decode %s: %v`, body, err)
	}
}

// decodeAuthenticatedJSON decodes a successful response into result, and fails the test
// unless the response reports the Fever API version and a successful authentication.
func decodeAuthenticatedJSON(t *testing.T, status int, body []byte, result any) {
	t.Helper()

	var base baseResult
	decodeJSON(t, status, body, &base)
	if base.APIVersion != feverAPIVersion || base.Auth != 1 {
		t.Fatalf(`Expected an authenticated response for API version %d, got %s`, feverAPIVersion, body)
	}
	decodeJSON(t, status, body, result)
}

// requireAuthFailure fails the test unless the response is the Fever authentication failure:
// HTTP 200, because Fever clients only look at the auth member, with auth set to 0.
func requireAuthFailure(t *testing.T, status int, body []byte) {
	t.Helper()

	var result baseResult
	decodeJSON(t, status, body, &result)
	if result.APIVersion != feverAPIVersion || result.Auth != 0 {
		t.Fatalf(`Expected an authentication failure for API version %d, got %s`, feverAPIVersion, body)
	}
}

// The response types below mirror the JSON documents sent to clients. They are not the
// package's own types, so renaming a JSON field fails the tests as it would break clients.

type baseResult struct {
	APIVersion          int   `json:"api_version"`
	Auth                int   `json:"auth"`
	LastRefreshedOnTime int64 `json:"last_refreshed_on_time"`
}
