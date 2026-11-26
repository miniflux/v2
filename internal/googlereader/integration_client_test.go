// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

const starredStream = "user/-/state/com.google/starred"

func feedStream(feedID int64) string {
	return "feed/" + strconv.FormatInt(feedID, 10)
}

// labelStream returns the stream ID of a label in the "user/-/" form, which clients send for the current user.
func labelStream(title string) string {
	return "user/-/label/" + title
}

// googleReaderClient calls the Google Reader API. It sends no credentials when token is empty.
type googleReaderClient struct {
	baseURL string
	token   string
}

// get sends a GET request with the token in the Authorization header.
func (client *googleReaderClient) get(t *testing.T, path string, query url.Values) (int, []byte) {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, client.baseURL+path+"?"+query.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if client.token != "" {
		request.Header.Set("Authorization", "GoogleLogin auth="+client.token)
	}
	return send(t, request)
}

// post sends a form with the token in the T field.
func (client *googleReaderClient) post(t *testing.T, path string, form url.Values) (int, []byte) {
	t.Helper()

	values := url.Values{}
	for key, list := range form {
		values[key] = list
	}
	if client.token != "" {
		values.Set("T", client.token)
	}

	request, err := http.NewRequest(http.MethodPost, client.baseURL+path, strings.NewReader(values.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return send(t, request)
}

// getJSON sends a GET request with output=json and decodes the successful response into result.
func (client *googleReaderClient) getJSON(t *testing.T, path string, query url.Values, result any) {
	t.Helper()

	values := url.Values{"output": {"json"}}
	for key, list := range query {
		values[key] = list
	}
	status, body := client.get(t, path, values)
	decodeJSON(t, status, body, result)
}

// postJSON sends a form with output=json and decodes the successful response into result.
func (client *googleReaderClient) postJSON(t *testing.T, path string, form url.Values, result any) {
	t.Helper()

	values := url.Values{"output": {"json"}}
	for key, list := range form {
		values[key] = list
	}
	status, body := client.post(t, path, values)
	decodeJSON(t, status, body, result)
}

// postOK sends a form and requires the plain text OK response.
func (client *googleReaderClient) postOK(t *testing.T, path string, form url.Values) {
	t.Helper()

	status, body := client.post(t, path, form)
	if status != http.StatusOK || string(body) != "OK" {
		t.Fatalf(`POST %s %v: expected status 200 with body "OK", got %d: %s`, path, form, status, body)
	}
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

// requireStatus fails the test when the response status is not the expected one.
func requireStatus(t *testing.T, status int, body []byte, expected int) {
	t.Helper()

	if status != expected {
		t.Fatalf(`Expected status %d, got %d: %s`, expected, status, body)
	}
}

// The response types below mirror the JSON documents sent to clients. They are not the
// package's own types, so renaming a JSON field fails the tests as it would break clients.

type loginResult struct {
	SID  string `json:"SID"`
	LSID string `json:"LSID"`
	Auth string `json:"Auth"`
}

type userInfoResult struct {
	UserID        string `json:"userId"`
	UserName      string `json:"userName"`
	UserProfileID string `json:"userProfileId"`
	UserEmail     string `json:"userEmail"`
}

type tagResult struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type subscriptionResult struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Categories []tagResult `json:"categories"`
	URL        string      `json:"url"`
	HTMLURL    string      `json:"htmlUrl"`
	IconURL    string      `json:"iconUrl"`
}

type subscriptionListResult struct {
	Subscriptions []subscriptionResult `json:"subscriptions"`
}

type quickAddResult struct {
	NumResults int64  `json:"numResults"`
	Query      string `json:"query"`
	StreamID   string `json:"streamId"`
	StreamName string `json:"streamName"`
}
