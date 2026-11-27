// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever_test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"os"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
	"miniflux.app/v2/internal/model"
)

const (
	skipIntegrationTestsMessage = `Set TEST_DATABASE_URL to a PostgreSQL database to run the Fever API integration tests`

	// testUserPassword is the Miniflux password of the test users, used by the REST API.
	testUserPassword = "miniflux_test_password"

	// testFeverPassword is the Fever password of the test users, hashed into their API key.
	testFeverPassword = "fever_test_password"
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

// testUser is a Miniflux user with the Fever API enabled.
type testUser struct {
	id       int64
	username string

	// fever calls the Fever API with the API key of the user.
	fever *feverClient

	// rest calls the REST API, to set up data and to check changes independently.
	rest *miniflux.Client
}

// newTestUser creates a user with the Fever API enabled that is removed when the test ends.
func newTestUser(t *testing.T) *testUser {
	t.Helper()

	server := requireTestServer(t)
	user, err := server.Store.CreateUser(&model.UserCreationRequest{
		Username: fmt.Sprintf("fever_test_user_%d", rand.Int()),
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

	apiKey := setFeverIntegration(t, user.ID, true, testFeverPassword)

	return &testUser{
		id:       user.ID,
		username: user.Username,
		fever:    &feverClient{baseURL: server.URL, apiKey: apiKey},
		rest:     miniflux.NewClient(server.URL, user.Username, testUserPassword),
	}
}

// setFeverIntegration changes the Fever settings of the user, using the Miniflux username,
// and returns the API key. No API can change these settings, so they are written to the store.
func setFeverIntegration(t *testing.T, userID int64, enabled bool, password string) string {
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

	apiKey := feverAPIKey(user.Username, password)
	integration.FeverEnabled = enabled
	integration.FeverUsername = user.Username
	integration.FeverToken = apiKey
	if err := server.Store.UpdateIntegration(integration); err != nil {
		t.Fatal(err)
	}
	return apiKey
}

// feverAPIKey returns the API key that Fever clients send: the MD5 hash of "username:password",
// which is also the token that the integrations page stores.
func feverAPIKey(username, password string) string {
	sum := md5.Sum([]byte(username + ":" + password))
	return hex.EncodeToString(sum[:])
}

// subscribeToFixture subscribes the user to a new fixture feed and returns the feed ID with its entries.
func subscribeToFixture(t *testing.T, user *testUser) (int64, miniflux.Entries) {
	t.Helper()

	feedID := servertest.NewFeedFixture(t).Subscribe(t, user.rest)
	return feedID, feedEntries(t, user, feedID)
}

// createCategory creates a category, which the Fever API calls a group.
func createCategory(t *testing.T, user *testUser, title string) *miniflux.Category {
	t.Helper()

	category, err := user.rest.CreateCategory(title)
	if err != nil {
		t.Fatal(err)
	}
	return category
}

// subscribeInCategory subscribes the user to a new fixture feed in the category and returns the feed ID.
func subscribeInCategory(t *testing.T, user *testUser, categoryID int64) int64 {
	t.Helper()

	feedID, err := user.rest.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    servertest.NewFeedFixture(t).FeedURL,
		CategoryID: categoryID,
	})
	if err != nil {
		t.Fatal(err)
	}
	return feedID
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
