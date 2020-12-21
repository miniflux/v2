// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package database // import "miniflux.app/database"

import (
	"database/sql"
	"fmt"

	// Postgresql driver import
	_ "github.com/lib/pq"

	"miniflux.app/logger"
)

// NewConnectionPool configures the database connection pool.
func NewConnectionPool(dsn string, minConnections, maxConnections int) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(minConnections)

	return db, nil
}

// Migrate executes database migrations.
func Migrate(db *sql.DB) {
	var currentVersion int
	db.QueryRow(`SELECT version FROM schema_version`).Scan(&currentVersion)

	fmt.Println("-> Current schema version:", currentVersion)
	fmt.Println("-> Latest schema version:", schemaVersion)

	for version := currentVersion; version < schemaVersion; version++ {
		newVersion := version + 1
		fmt.Println("* Migrating to version:", newVersion)

		tx, err := db.Begin()
		if err != nil {
			logger.Fatal("[Migration v%d] %v", newVersion, err)
		}

		if err := migrations[version](tx); err != nil {
			tx.Rollback()
			logger.Fatal("[Migration v%d] %v", newVersion, err)
		}

		if _, err := tx.Exec(`DELETE FROM schema_version`); err != nil {
			tx.Rollback()
			logger.Fatal("[Migration v%d] %v", newVersion, err)
		}

		if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES ($1)`, newVersion); err != nil {
			tx.Rollback()
			logger.Fatal("[Migration v%d] %v", newVersion, err)
		}

		if err := tx.Commit(); err != nil {
			logger.Fatal("[Migration v%d] %v", newVersion, err)
		}
	}
}

// IsSchemaUpToDate checks if the database schema is up to date.
func IsSchemaUpToDate(db *sql.DB) error {
	var currentVersion int
	db.QueryRow(`SELECT version FROM schema_version`).Scan(&currentVersion)
	if currentVersion < schemaVersion {
		return fmt.Errorf(`the database schema is not up to date: current=v%d expected=v%d`, currentVersion, schemaVersion)
	}
	return nil
}
