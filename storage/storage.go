// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"context"
	"database/sql"
	"time"
)

// Storage handles all operations related to the database.
type Storage struct {
	db *sql.DB
}

// NewStorage returns a new Storage.
func NewStorage(db *sql.DB) *Storage {
	return &Storage{db}
}

// DatabaseVersion returns the version of the database which is in use.
func (s *Storage) DatabaseVersion() string {
	var dbVersion string
	err := s.db.QueryRow(`SELECT current_setting('server_version')`).Scan(&dbVersion)
	if err != nil {
		return err.Error()
	}

	return dbVersion
}

// Ping checks if the database connection works.
func (s *Storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.db.PingContext(ctx)
}

// DBStats returns database statistics.
func (s *Storage) DBStats() sql.DBStats {
	return s.db.Stats()
}
