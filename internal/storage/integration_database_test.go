// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"miniflux.app/v2/internal/database"
)

const skipIntegrationTestsMessage = `Set TEST_DATABASE_URL to a PostgreSQL database to run the storage integration tests`

// testDatabaseURL is the database migrated by TestMain, or empty when TEST_DATABASE_URL is not set.
var testDatabaseURL string

func TestMain(m *testing.M) {
	if databaseURL := os.Getenv("TEST_DATABASE_URL"); databaseURL != "" {
		if err := migrateDatabase(databaseURL); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to migrate the integration test database: %v\n", err)
			os.Exit(1)
		}
		testDatabaseURL = databaseURL
	}

	os.Exit(m.Run())
}

func migrateDatabase(databaseURL string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	return database.Migrate(db)
}

// openTestDatabase connects to the database migrated by TestMain, or skips the test when TEST_DATABASE_URL is not set.
func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	if testDatabaseURL == "" {
		t.Skip(skipIntegrationTestsMessage)
	}

	db, err := sql.Open("postgres", testDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
