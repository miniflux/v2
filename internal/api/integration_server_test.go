// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
	"miniflux.app/v2/internal/model"
)

const skipIntegrationTestsMessage = `Set TEST_DATABASE_URL to a PostgreSQL database to run the API integration tests`

// integrationConfig describes the Miniflux server started by TestMain.
// It is nil when TEST_DATABASE_URL is not set.
var integrationConfig *integrationTestConfig

type integrationTestConfig struct {
	testBaseURL         string
	testAdminUsername   string
	testAdminPassword   string
	testRegularUsername string
	testRegularPassword string
}

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		return m.Run()
	}

	stopServer, err := startIntegrationTestServer(databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start the integration test server: %v\n", err)
		return 1
	}
	defer stopServer()

	return m.Run()
}

// startIntegrationTestServer starts a Miniflux server backed by the given database
// and creates its admin user.
func startIntegrationTestServer(databaseURL string) (func(), error) {
	testServer, err := servertest.NewServer(databaseURL)
	if err != nil {
		return nil, err
	}

	testConfig := &integrationTestConfig{
		testBaseURL:         testServer.URL,
		testAdminUsername:   "admin",
		testAdminPassword:   "test123",
		testRegularUsername: "regular_test_user",
		testRegularPassword: "regular_test_user_password",
	}

	if !testServer.Store.UserExists(testConfig.testAdminUsername) {
		if _, err := testServer.Store.CreateUser(&model.UserCreationRequest{
			Username: testConfig.testAdminUsername,
			Password: testConfig.testAdminPassword,
			IsAdmin:  true,
		}); err != nil {
			testServer.Close()
			return nil, err
		}
	}

	integrationConfig = testConfig
	return testServer.Close, nil
}

// requireIntegrationTestConfig returns the configuration of the integration test server,
// or skips the test when TEST_DATABASE_URL is not set.
func requireIntegrationTestConfig(t *testing.T) *integrationTestConfig {
	t.Helper()

	if integrationConfig == nil {
		t.Skip(skipIntegrationTestsMessage)
	}
	return integrationConfig
}

func (c *integrationTestConfig) genRandomUsername() string {
	return fmt.Sprintf("%s_%10d", c.testRegularUsername, rand.Int())
}

func (c *integrationTestConfig) newAdminClient() *miniflux.Client {
	return miniflux.NewClient(c.testBaseURL, c.testAdminUsername, c.testAdminPassword)
}

// deleteUserOnCleanup removes the user when the test and its subtests finish,
// and fails the test if the user cannot be removed.
func (c *integrationTestConfig) deleteUserOnCleanup(t *testing.T, userID int64) {
	t.Helper()

	t.Cleanup(func() {
		if err := c.newAdminClient().DeleteUser(userID); err != nil {
			t.Errorf(`Unable to delete user %d: %v`, userID, err)
		}
	})
}

// newRegularUser creates a regular user that is removed when the test ends,
// and returns a client authenticated as that user.
func (c *integrationTestConfig) newRegularUser(t *testing.T) (*miniflux.Client, *miniflux.User) {
	t.Helper()

	user, err := c.newAdminClient().CreateUser(c.genRandomUsername(), c.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	c.deleteUserOnCleanup(t, user.ID)

	return miniflux.NewClient(c.testBaseURL, user.Username, c.testRegularPassword), user
}

// requestStatus sends an API request authenticated as the regular user and returns the status code.
// It covers routes that the client does not use, such as aliases and alternative methods.
func (c *integrationTestConfig) requestStatus(t *testing.T, user *miniflux.User, method, path string) int {
	t.Helper()

	request, err := http.NewRequest(method, c.testBaseURL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.SetBasicAuth(user.Username, c.testRegularPassword)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	return response.StatusCode
}

// listFeedEntries returns the entries of the feed that match the filter.
func listFeedEntries(t *testing.T, client *miniflux.Client, feedID int64, filter *miniflux.Filter) *miniflux.EntryResultSet {
	t.Helper()

	result, err := client.FeedEntries(feedID, filter)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
