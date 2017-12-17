// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"fmt"

	"github.com/miniflux/miniflux/helper"
	"github.com/miniflux/miniflux/model"
)

// CreateSession creates a new session.
func (s *Storage) CreateSession() (*model.Session, error) {
	session := model.Session{
		ID:   helper.GenerateRandomString(32),
		Data: &model.SessionData{CSRF: helper.GenerateRandomString(64)},
	}

	query := "INSERT INTO sessions (id, data) VALUES ($1, $2)"
	_, err := s.db.Exec(query, session.ID, session.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to create session: %v", err)
	}

	return &session, nil
}

// UpdateSessionField updates only one session field.
func (s *Storage) UpdateSessionField(sessionID, field string, value interface{}) error {
	query := `UPDATE sessions
		SET data = jsonb_set(data, '{%s}', to_jsonb($1::text), true)
		WHERE id=$2`

	_, err := s.db.Exec(fmt.Sprintf(query, field), value, sessionID)
	if err != nil {
		return fmt.Errorf("unable to update session field: %v", err)
	}

	return nil
}

// Session returns the given session.
func (s *Storage) Session(id string) (*model.Session, error) {
	var session model.Session

	query := "SELECT id, data FROM sessions WHERE id=$1"
	err := s.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.Data,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch session: %v", err)
	}

	return &session, nil
}

// FlushAllSessions removes all sessions from the database.
func (s *Storage) FlushAllSessions() (err error) {
	_, err = s.db.Exec(`DELETE FROM user_sessions`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`DELETE FROM sessions`)
	if err != nil {
		return err
	}

	return nil
}

// CleanOldSessions removes sessions older than 30 days.
func (s *Storage) CleanOldSessions() int64 {
	query := `DELETE FROM sessions
		WHERE id IN (SELECT id FROM sessions WHERE created_at < now() - interval '30 days')`

	result, err := s.db.Exec(query)
	if err != nil {
		return 0
	}

	n, _ := result.RowsAffected()
	return n
}
