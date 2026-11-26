// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader_test

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestClientLogin(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	anonymous := &googleReaderClient{baseURL: requireTestServer(t).URL}
	credentials := url.Values{"Email": {user.username}, "Passwd": {testGoogleReaderPassword}}

	if !strings.HasPrefix(user.reader.token, user.username+"/") {
		t.Errorf(`Expected the token to start with the username, got %q`, user.reader.token)
	}

	t.Run("json output", func(t *testing.T) {
		var result loginResult
		anonymous.postJSON(t, "/accounts/ClientLogin", credentials, &result)

		expected := loginResult{SID: user.reader.token, LSID: user.reader.token, Auth: user.reader.token}
		if result != expected {
			t.Errorf(`Unexpected response, got %+v instead of %+v`, result, expected)
		}
	})

	t.Run("text output", func(t *testing.T) {
		status, body := anonymous.post(t, "/accounts/ClientLogin", credentials)
		requireStatus(t, status, body, http.StatusOK)

		expected := fmt.Sprintf("SID=%[1]s\nLSID=%[1]s\nAuth=%[1]s\n", user.reader.token)
		if string(body) != expected {
			t.Errorf(`Unexpected response, got %q instead of %q`, body, expected)
		}
	})
}

func TestClientLoginRejectsInvalidCredentials(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	disabledUser := newTestUser(t)
	setGoogleReaderIntegration(t, disabledUser.id, false, testGoogleReaderPassword)

	anonymous := &googleReaderClient{baseURL: requireTestServer(t).URL}

	scenarios := []struct {
		name     string
		username string
		password string
	}{
		{"wrong password", user.username, "wrong password"},
		{"Miniflux password", user.username, testUserPassword},
		{"unknown username", "unknown_googlereader_user", testGoogleReaderPassword},
		{"empty username", "", testGoogleReaderPassword},
		{"empty password", user.username, ""},
		{"Google Reader disabled", disabledUser.username, testGoogleReaderPassword},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			status, body := anonymous.post(t, "/accounts/ClientLogin", url.Values{
				"Email":  {scenario.username},
				"Passwd": {scenario.password},
				"output": {"json"},
			})
			requireStatus(t, status, body, http.StatusUnauthorized)
		})
	}
}

func TestAuthenticationRejectsInvalidTokens(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	otherUser := newTestUser(t)
	baseURL := requireTestServer(t).URL

	_, hash, _ := strings.Cut(user.reader.token, "/")
	_, otherHash, _ := strings.Cut(otherUser.reader.token, "/")
	tamperedToken := user.username + "/" + strings.Repeat("0", len(hash))

	// Malformed credentials are covered by the middleware unit tests: these tokens are well-formed.
	getScenarios := []struct {
		name  string
		token string
	}{
		{"tampered hash", tamperedToken},
		{"hash of another user", user.username + "/" + otherHash},
		{"unknown username", "unknown_googlereader_user/" + hash},
	}

	for _, scenario := range getScenarios {
		t.Run("GET with "+scenario.name, func(t *testing.T) {
			client := &googleReaderClient{baseURL: baseURL, token: scenario.token}
			status, body := client.get(t, "/reader/api/0/user-info", url.Values{"output": {"json"}})
			requireStatus(t, status, body, http.StatusUnauthorized)
		})
	}

	// POST requests are authenticated by the T field only, the Authorization header is ignored.
	postScenarios := []struct {
		name          string
		token         string
		authorization string
	}{
		{"no T field but a valid Authorization header", "", "GoogleLogin auth=" + user.reader.token},
		{"tampered T field", tamperedToken, ""},
		{"T field of another user", user.username + "/" + otherHash, ""},
	}

	for _, scenario := range postScenarios {
		t.Run("POST with "+scenario.name, func(t *testing.T) {
			form := url.Values{"output": {"json"}}
			if scenario.token != "" {
				form.Set("T", scenario.token)
			}

			request, err := http.NewRequest(http.MethodPost, baseURL+"/reader/api/0/stream/items/contents", strings.NewReader(form.Encode()))
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if scenario.authorization != "" {
				request.Header.Set("Authorization", scenario.authorization)
			}

			status, body := send(t, request)
			requireStatus(t, status, body, http.StatusUnauthorized)
		})
	}
}

func TestTokenIsRevokedWhenGoogleReaderSettingsChange(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name     string
		enabled  bool
		password string
	}{
		{"password changed", true, "new_googlereader_password"},
		{"integration disabled", false, testGoogleReaderPassword},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			user := newTestUser(t)
			setGoogleReaderIntegration(t, user.id, scenario.enabled, scenario.password)

			status, body := user.reader.get(t, "/reader/api/0/user-info", url.Values{"output": {"json"}})
			requireStatus(t, status, body, http.StatusUnauthorized)
		})
	}
}

func TestTokenEndpoint(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	status, body := user.reader.get(t, "/reader/api/0/token", nil)
	requireStatus(t, status, body, http.StatusOK)

	if string(body) != user.reader.token {
		t.Errorf(`Unexpected token, got %q instead of %q`, body, user.reader.token)
	}
}

func TestUserInfoEndpoint(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	var result userInfoResult
	user.reader.getJSON(t, "/reader/api/0/user-info", nil, &result)

	userID := strconv.FormatInt(user.id, 10)
	expected := userInfoResult{UserID: userID, UserName: user.username, UserProfileID: userID, UserEmail: user.username}
	if result != expected {
		t.Errorf(`Unexpected user info, got %+v instead of %+v`, result, expected)
	}
}

func TestUnimplementedEndpointsReturnEmptyList(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	anonymous := &googleReaderClient{baseURL: requireTestServer(t).URL}

	status, body := user.reader.get(t, "/reader/api/0/unread-count", nil)
	requireStatus(t, status, body, http.StatusOK)
	if strings.TrimSpace(string(body)) != "[]" {
		t.Errorf(`Expected an empty list for GET, got %s`, body)
	}

	status, body = user.reader.post(t, "/reader/api/0/subscription/export", nil)
	requireStatus(t, status, body, http.StatusOK)
	if strings.TrimSpace(string(body)) != "[]" {
		t.Errorf(`Expected an empty list for POST, got %s`, body)
	}

	status, body = anonymous.get(t, "/reader/api/0/unread-count", nil)
	requireStatus(t, status, body, http.StatusUnauthorized)
}
