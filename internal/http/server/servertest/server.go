// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package servertest runs an in-process Miniflux server for integration tests.
package servertest // import "miniflux.app/v2/internal/http/server/servertest"

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http/httptest"
	"os"
	"testing"

	"miniflux.app/v2/internal/cli"
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/database"
	"miniflux.app/v2/internal/http/server"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/worker"
)

const skipIntegrationTestsMessage = `Set TEST_DATABASE_URL to a PostgreSQL database to run the integration tests`

// testServer is the server started by Main, or nil when TEST_DATABASE_URL is not set.
var testServer *Server

// Server serves all Miniflux routes from an httptest server backed by a PostgreSQL database.
type Server struct {
	// URL is the base URL of the server, without a trailing slash.
	URL string

	// Store gives the tests direct access to the database, for setup that no API offers.
	Store *storage.Storage

	db         *sql.DB
	pool       *worker.Pool
	httpServer *httptest.Server
}

// Main runs the tests of a package and exits. When TEST_DATABASE_URL is set, it first starts
// the server that Require returns, and prepares it with setup unless setup is nil.
func Main(m *testing.M, setup func(*Server) error) {
	os.Exit(runTests(m, setup))
}

func runTests(m *testing.M, setup func(*Server) error) int {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		return m.Run()
	}

	server, err := newServer(databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start the integration test server: %v\n", err)
		return 1
	}
	defer server.stop()

	if setup != nil {
		if err := setup(server); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to set up the integration test server: %v\n", err)
			return 1
		}
	}

	testServer = server
	return m.Run()
}

// Require returns the server started by Main, or skips the test when TEST_DATABASE_URL is not set.
func Require(t testing.TB) *Server {
	t.Helper()

	if testServer == nil {
		t.Skip(skipIntegrationTestsMessage)
	}
	return testServer
}

// NewUser creates a regular user with the given password, removed when the test ends.
func (s *Server) NewUser(t testing.TB, password string) *model.User {
	t.Helper()

	user, err := s.Store.CreateUser(&model.UserCreationRequest{
		Username: fmt.Sprintf("test_user_%d", rand.Int()),
		Password: password,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Store.RemoveUser(user.ID); err != nil {
			t.Errorf(`Unable to remove user %d: %v`, user.ID, err)
		}
	})
	return user
}

// newServer migrates the database and starts a server that uses it.
func newServer(databaseURL string) (*Server, error) {
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

	store := storage.NewStorage(db)
	pool := worker.NewPool(store, opts.WorkerPoolSize())
	httpServer := httptest.NewServer(server.NewRouter(store, pool))

	return &Server{
		URL:        httpServer.URL,
		Store:      store,
		db:         db,
		pool:       pool,
		httpServer: httpServer,
	}, nil
}

// stop stops the server and closes the database.
func (s *Server) stop() {
	s.httpServer.Close()
	s.pool.Shutdown()
	s.db.Close()
}
