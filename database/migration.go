// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package database // import "miniflux.app/database"

import (
	"database/sql"
	"fmt"
	"strconv"

	"miniflux.app/logger"
)

const schemaVersion = 25

// Migrate executes database migrations.
func Migrate(db *sql.DB) {
	var currentVersion int
	db.QueryRow(`select version from schema_version`).Scan(&currentVersion)

	fmt.Println("Current schema version:", currentVersion)
	fmt.Println("Latest schema version:", schemaVersion)

	for version := currentVersion + 1; version <= schemaVersion; version++ {
		fmt.Println("Migrating to version:", version)

		tx, err := db.Begin()
		if err != nil {
			logger.Fatal("[Migrate] %v", err)
		}

		rawSQL := SqlMap["schema_version_"+strconv.Itoa(version)]
		// fmt.Println(rawSQL)
		_, err = tx.Exec(rawSQL)
		if err != nil {
			tx.Rollback()
			logger.Fatal("[Migrate] %v", err)
		}

		if _, err := tx.Exec(`delete from schema_version`); err != nil {
			tx.Rollback()
			logger.Fatal("[Migrate] %v", err)
		}

		if _, err := tx.Exec(`insert into schema_version (version) values($1)`, version); err != nil {
			tx.Rollback()
			logger.Fatal("[Migrate] %v", err)
		}

		if err := tx.Commit(); err != nil {
			logger.Fatal("[Migrate] %v", err)
		}
	}
}
