// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader_test

import (
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/server/servertest"
	"miniflux.app/v2/internal/model"
)

const (
	skipIntegrationTestsMessage = `Set TEST_DATABASE_URL to a PostgreSQL database to run the Google Reader API integration tests`

	// testUserPassword is the Miniflux password of the test users, used by the REST API.
	testUserPassword = "miniflux_test_password"

	// testGoogleReaderPassword is the Google Reader password of the test users, used by ClientLogin.
	testGoogleReaderPassword = "googlereader_test_password"
)

// testServer is the Miniflux server started by TestMain, or nil when TEST_DATABASE_URL is not set.
var testServer *servertest.Server

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		return m.Run()
	}

	server, err := servertest.NewServer(databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start the integration test server: %v\n", err)
		return 1
	}
	defer server.Close()

	testServer = server
	return m.Run()
}

// requireTestServer returns the integration test server, or skips the test when TEST_DATABASE_URL is not set.
func requireTestServer(t *testing.T) *servertest.Server {
	t.Helper()

	if testServer == nil {
		t.Skip(skipIntegrationTestsMessage)
	}
	return testServer
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

	server := requireTestServer(t)
	user, err := server.Store.CreateUser(&model.UserCreationRequest{
		Username: fmt.Sprintf("googlereader_test_user_%d", rand.Int()),
		Password: testUserPassword,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := server.Store.RemoveUser(user.ID); err != nil {
			t.Errorf(`Unable to remove user %d: %v`, user.ID, err)
		}
	})

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

	server := requireTestServer(t)
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
	anonymous := &googleReaderClient{baseURL: requireTestServer(t).URL}
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
	return feedID, feedEntries(t, user, feedID)
}

// feedEntries returns the entries of the feed oldest first: entry 1 is published on 2025-01-01, entry 3 on 2025-01-03.
func feedEntries(t *testing.T, user *testUser, feedID int64) miniflux.Entries {
	t.Helper()

	result, err := user.rest.FeedEntries(feedID, &miniflux.Filter{Order: "published_at", Direction: "asc"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != servertest.EntryCount {
		t.Fatalf(`Expected %d entries, got %d`, servertest.EntryCount, len(result.Entries))
	}
	return result.Entries
}
