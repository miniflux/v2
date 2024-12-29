//go:build sqlite

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package database // import "miniflux.app/v2/internal/database"

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// NewConnectionPool configures the database connection pool.
func NewConnectionPool(dsn string, _, _ int, _ time.Duration) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getDriverStr() string {
	return "sqlite3"
}
