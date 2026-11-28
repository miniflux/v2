// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader_test

import (
	"net/url"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/server/servertest"
)

const (
	// testUserPassword is the Miniflux password of the test users, used by the REST API.
	testUserPassword = "miniflux_test_password"

	// testGoogleReaderPassword is the Google Reader password of the test users, used by ClientLogin.
	testGoogleReaderPassword = "googlereader_test_password"
)

func TestMain(m *testing.M) {
	servertest.Main(m, nil)
}

// testUser is a Miniflux user with the Google Reader API enabled.
type testUser struct {
	id       int64
	username string

	// reader calls the Google Reader API with the token returned by ClientLogin.
	reader *googleReaderClient

	// rest calls the REST API, to set up data and to check changes independently.
	rest *miniflux.Client
}

// newTestUser creates a user with the Google Reader API enabled that is removed when the test ends.
func newTestUser(t *testing.T) *testUser {
	t.Helper()

	server := servertest.Require(t)
	user := server.NewUser(t, testUserPassword)
	setGoogleReaderIntegration(t, user.ID, true, testGoogleReaderPassword)

	return &testUser{
		id:       user.ID,
		username: user.Username,
		reader:   &googleReaderClient{baseURL: server.URL, token: clientLogin(t, user.Username, testGoogleReaderPassword)},
		rest:     miniflux.NewClient(server.URL, user.Username, testUserPassword),
	}
}

// setGoogleReaderIntegration changes the Google Reader settings of the user, using the Miniflux username.
func setGoogleReaderIntegration(t *testing.T, userID int64, enabled bool, password string) {
	t.Helper()

	server := servertest.Require(t)
	integration, err := server.Store.Integration(userID)
	if err != nil {
		t.Fatal(err)
	}

	user, err := server.Store.UserByID(userID)
	if err != nil {
		t.Fatal(err)
	}

	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	integration.GoogleReaderEnabled = enabled
	integration.GoogleReaderUsername = user.Username
	integration.GoogleReaderPassword = passwordHash
	if err := server.Store.UpdateIntegration(integration); err != nil {
		t.Fatal(err)
	}
}

// clientLogin authenticates with ClientLogin and returns the auth token.
func clientLogin(t *testing.T, username, password string) string {
	t.Helper()

	var result loginResult
	anonymous := &googleReaderClient{baseURL: servertest.Require(t).URL}
	anonymous.postJSON(t, "/accounts/ClientLogin", url.Values{
		"Email":  {username},
		"Passwd": {password},
	}, &result)

	if result.Auth == "" {
		t.Fatalf(`ClientLogin returned an empty token: %+v`, result)
	}
	return result.Auth
}

// subscribeToFixture subscribes the user to a new fixture feed and returns the feed ID with its entries.
func subscribeToFixture(t *testing.T, user *testUser) (int64, miniflux.Entries) {
	t.Helper()

	feedID := servertest.NewFeedFixture(t).Subscribe(t, user.rest)
	return feedID, servertest.FeedEntries(t, user.rest, feedID)
}
