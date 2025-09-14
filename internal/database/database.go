// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package database // import "miniflux.app/v2/internal/database"

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type DBKind int

const (
	DBKindPostgres DBKind = iota
	DBKindCockroach
	DBKindSqlite
)

var dbKindProto = map[DBKind]string{
	DBKindPostgres:  "postgresql",
	DBKindCockroach: "cockroachdb",
	DBKindSqlite:    "sqlite",
}

var dbKindDriver = map[DBKind]string{
	DBKindPostgres:  "postgres",
	DBKindCockroach: "postgres",
	DBKindSqlite:    "sqlite",
}

func DetectKind(conn string) (DBKind, error) {
	switch {
	case strings.HasPrefix(conn, "postgres"),
		strings.HasPrefix(conn, "postgresql"):
		return DBKindPostgres, nil
	case strings.HasPrefix(conn, "cockroach"),
		strings.HasPrefix(conn, "cockroachdb"):
		return DBKindCockroach, nil
	case strings.HasPrefix(conn, "file"),
		strings.HasPrefix(conn, "sqlite"):
		return DBKindSqlite, nil
	default:
		return 0, fmt.Errorf("unknown db kind in conn string: %q", conn)
	}
}

type Migration func(*sql.Tx) error

var dbKindMigrations = map[DBKind][]Migration{
	DBKindPostgres:  postgresMigrations,
	DBKindCockroach: cockroachMigrations,
	DBKindSqlite:    sqliteMigrations,
}

var dbKindSchemaVersion = map[DBKind]int{
	DBKindPostgres:  postgresSchemaVersion,
	DBKindCockroach: cockroachSchemaVersion,
	DBKindSqlite:    sqliteSchemaVersion,
}

// Migrate executes database migrations.
func Migrate(kind DBKind, db *sql.DB) error {
	var currentVersion int
	db.QueryRow(`SELECT version FROM schema_version`).Scan(&currentVersion)

	migrations := dbKindMigrations[kind]
	schemaVersion := dbKindSchemaVersion[kind]

	slog.Info("Running database migrations",
		slog.Int("current_version", currentVersion),
		slog.Int("latest_version", schemaVersion),
	)

	for version := currentVersion; version < schemaVersion; version++ {
		newVersion := version + 1

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if err := migrations[version](tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if kind == DBKindSqlite {
			if _, err := tx.Exec(`DELETE FROM schema_version`); err != nil {
				tx.Rollback()
				return fmt.Errorf("[Migration v%d] %v", newVersion, err)
			}
			if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES (?)`, newVersion); err != nil {
				tx.Rollback()
				return fmt.Errorf("[Migration v%d] %v", newVersion, err)
			}
		} else {
			if _, err := tx.Exec(`TRUNCATE schema_version`); err != nil {
				tx.Rollback()
				return fmt.Errorf("[Migration v%d] %v", newVersion, err)
			}
			if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES ($1)`, newVersion); err != nil {
				tx.Rollback()
				return fmt.Errorf("[Migration v%d] %v", newVersion, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}
	}

	return nil
}

// IsSchemaUpToDate checks if the database schema is up to date.
func IsSchemaUpToDate(kind DBKind, db *sql.DB) error {
	schemaVersion := dbKindSchemaVersion[kind]

	var currentVersion int
	db.QueryRow(`SELECT version FROM schema_version`).Scan(&currentVersion)
	if currentVersion < schemaVersion {
		return fmt.Errorf(`the database schema is not up to date: current=v%d expected=v%d`, currentVersion, schemaVersion)
	}
	return nil
}

func NewConnectionPool(kind DBKind, dsn string, minConnections, maxConnections int, connectionLifetime time.Duration) (*sql.DB, error) {
	driver := dbKindDriver[kind]

	// replace cockroachdb protocol with postgres
	// we use cockroachdb protocol to detect cockroachdb but go wants postgres
	if kind == DBKindCockroach {
		split := strings.SplitN(dsn, ":", 2)
		dsn = fmt.Sprintf("postgres:%s", split[1])
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(minConnections)
	db.SetConnMaxLifetime(connectionLifetime)

	return db, nil
}
