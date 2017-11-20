// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) Close() {
	s.db.Close()
}

func NewStorage(databaseUrl string, maxOpenConns int) *Storage {
	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(2)

	return &Storage{db: db}
}
