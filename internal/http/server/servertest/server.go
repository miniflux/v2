// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package servertest runs an in-process Miniflux server for integration tests.
package servertest // import "miniflux.app/v2/internal/http/server/servertest"

import (
	"database/sql"
	"log/slog"
	"net/http/httptest"
	"os"

	"miniflux.app/v2/internal/cli"
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/database"
	"miniflux.app/v2/internal/http/server"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/worker"
)

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

// NewServer migrates the database and starts a server that uses it.
func NewServer(databaseURL string) (*Server, error) {
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

// Close stops the server and closes the database.
func (s *Server) Close() {
	s.httpServer.Close()
	s.pool.Shutdown()
	s.db.Close()
}
