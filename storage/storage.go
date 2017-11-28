// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"log"

	// Postgresql driver import
	_ "github.com/lib/pq"
)

// Storage handles all operations related to the database.
type Storage struct {
	db *sql.DB
}

// Close closes all database connections.
func (s *Storage) Close() {
	s.db.Close()
}

// NewStorage returns a new Storage.
func NewStorage(databaseURL string, maxOpenConns int) *Storage {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(2)

	return &Storage{db: db}
}
