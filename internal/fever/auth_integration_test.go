// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever_test

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
)

func TestAuthentication(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	t.Run("API key in the query string", func(t *testing.T) {
		before := time.Now().Unix()
		var result baseResult
		user.fever.getJSON(t, nil, &result)
		after := time.Now().Unix()

		// Miniflux reports the time of the response, not the time of the last feed refresh.
		if result.LastRefreshedOnTime < before || result.LastRefreshedOnTime > after {
			t.Errorf(`Expected last_refreshed_on_time between %d and %d, got %d`, before, after, result.LastRefreshedOnTime)
		}
	})

	t.Run("API key in the form body", func(t *testing.T) {
		var result baseResult
		user.fever.postJSON(t, nil, &result)
	})

	t.Run("API key in uppercase", func(t *testing.T) {
		client := &feverClient{baseURL: requireTestServer(t).URL, apiKey: strings.ToUpper(user.fever.apiKey)}

		var result baseResult
		client.getJSON(t, nil, &result)
	})

	t.Run("selector without a value", func(t *testing.T) {
		status, body := user.fever.getRawQuery(t, "groups&api_key="+user.fever.apiKey)

		var result map[string]json.RawMessage
		decodeAuthenticatedJSON(t, status, body, &result)
		if _, found := result["groups"]; !found {
			t.Errorf(`Expected the groups member, got %s`, body)
		}
	})
}

func TestAuthenticationRejectsInvalidAPIKeys(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	disabledUser := newTestUser(t)
	disabledAPIKey := setFeverIntegration(t, disabledUser.id, false, testFeverPassword)

	// The anonymous client sends the API key of each scenario as is.
	anonymous := &feverClient{baseURL: requireTestServer(t).URL}

	scenarios := []struct {
		name   string
		params url.Values
	}{
		{"no API key", nil},
		{"an empty API key", url.Values{"api_key": {""}}},
		{"the key of a wrong password", url.Values{"api_key": {feverAPIKey(user.username, "wrong password")}}},
		{"the key of the Miniflux password", url.Values{"api_key": {feverAPIKey(user.username, testUserPassword)}}},
		{"the key of an unknown username", url.Values{"api_key": {feverAPIKey("unknown_fever_user", testFeverPassword)}}},
		{"the key of a user with Fever disabled", url.Values{"api_key": {disabledAPIKey}}},
	}

	for _, scenario := range scenarios {
		t.Run("GET with "+scenario.name, func(t *testing.T) {
			status, body := anonymous.get(t, scenario.params)
			requireAuthFailure(t, status, body)
		})

		t.Run("POST with "+scenario.name, func(t *testing.T) {
			status, body := anonymous.post(t, scenario.params)
			requireAuthFailure(t, status, body)
		})
	}
}

func TestRejectedRequestsDoNotChangeEntries(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	_, entries := subscribeToFixture(t, user)
	entryID := entries[0].ID
	markAsRead := url.Values{"mark": {"item"}, "as": {"read"}, "id": {strconv.FormatInt(entryID, 10)}}

	intruder := &feverClient{baseURL: requireTestServer(t).URL, apiKey: feverAPIKey(user.username, "wrong password")}
	status, body := intruder.post(t, markAsRead)
	requireAuthFailure(t, status, body)
	requireEntryStatus(t, user, entryID, miniflux.EntryStatusUnread)

	// The same request with the right API key changes the entry, so the rejection above is what kept it unread.
	var result baseResult
	user.fever.postJSON(t, markAsRead, &result)
	requireEntryStatus(t, user, entryID, miniflux.EntryStatusRead)
}

// requireEntryStatus fails the test unless the REST API reports the expected status for the entry.
func requireEntryStatus(t *testing.T, user *testUser, entryID int64, expected string) {
	t.Helper()

	entry, err := user.rest.Entry(entryID)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Status != expected {
		t.Fatalf(`Expected entry %d to be %q, got %q`, entryID, expected, entry.Status)
	}
}
