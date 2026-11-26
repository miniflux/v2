// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http/httptest"
	"os"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/cli"
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/database"
	"miniflux.app/v2/internal/http/server"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/worker"
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

// startIntegrationTestServer serves all Miniflux routes from an in-process HTTP server
// backed by the given database, which is migrated and given an admin user.
func startIntegrationTestServer(databaseURL string) (func(), error) {
	// The tests serve their feeds from httptest servers listening on 127.0.0.1.
	os.Setenv("FETCHER_ALLOW_PRIVATE_NETWORKS", "1")
	os.Setenv("INTEGRATION_ALLOW_PRIVATE_NETWORKS", "1")

	opts, err := config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		return nil, err
	}
	config.Opts = opts

	// Keep the server logs out of the test output unless LOG_LEVEL is set.
	if os.Getenv("LOG_LEVEL") == "" {
		slog.SetDefault(slog.New(slog.DiscardHandler))
	} else if err := cli.InitializeDefaultLogger(opts.LogLevel(), os.Stderr, opts.LogFormat(), opts.LogDateTime()); err != nil {
		return nil, err
	}

	db, err := database.NewConnectionPool(databaseURL, opts.DatabaseMinConns(), opts.DatabaseMaxConns(), opts.DatabaseConnectionLifetime())
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	testConfig := &integrationTestConfig{
		testAdminUsername:   "admin",
		testAdminPassword:   "test123",
		testRegularUsername: "regular_test_user",
		testRegularPassword: "regular_test_user_password",
	}

	store := storage.NewStorage(db)
	if !store.UserExists(testConfig.testAdminUsername) {
		if _, err := store.CreateUser(&model.UserCreationRequest{
			Username: testConfig.testAdminUsername,
			Password: testConfig.testAdminPassword,
			IsAdmin:  true,
		}); err != nil {
			db.Close()
			return nil, err
		}
	}

	pool := worker.NewPool(store, opts.WorkerPoolSize())
	httpServer := httptest.NewServer(server.NewRouter(store, pool))
	testConfig.testBaseURL = httpServer.URL
	integrationConfig = testConfig

	return func() {
		httpServer.Close()
		pool.Shutdown()
		db.Close()
	}, nil
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
