// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"fmt"

	"miniflux.app/crypto"
	"miniflux.app/model"
)

// CreateAppSessionWithUserPrefs creates a new application session with the given user preferences.
func (s *Storage) CreateAppSessionWithUserPrefs(userID int64) (*model.Session, error) {
	user, err := s.UserByID(userID)
	if err != nil {
		return nil, err
	}

	session := model.Session{
		ID: crypto.GenerateRandomString(32),
		Data: &model.SessionData{
			CSRF:     crypto.GenerateRandomString(64),
			Theme:    user.Theme,
			Language: user.Language,
		},
	}

	return s.createAppSession(&session)
}

// CreateAppSession creates a new application session.
func (s *Storage) CreateAppSession() (*model.Session, error) {
	session := model.Session{
		ID: crypto.GenerateRandomString(32),
		Data: &model.SessionData{
			CSRF: crypto.GenerateRandomString(64),
		},
	}

	return s.createAppSession(&session)
}

func (s *Storage) createAppSession(session *model.Session) (*model.Session, error) {
	_, err := s.db.Exec(`INSERT INTO sessions (id, data) VALUES ($1, $2)`, session.ID, session.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to create app session: %v", err)
	}

	return session, nil
}

// UpdateAppSessionField updates only one session field.
func (s *Storage) UpdateAppSessionField(sessionID, field string, value interface{}) error {
	query := `UPDATE sessions
		SET data = jsonb_set(data, '{%s}', to_jsonb($1::text), true)
		WHERE id=$2`

	_, err := s.db.Exec(fmt.Sprintf(query, field), value, sessionID)
	if err != nil {
		return fmt.Errorf("unable to update session field: %v", err)
	}

	return nil
}

// AppSession returns the given session.
func (s *Storage) AppSession(id string) (*model.Session, error) {
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

// CleanOldSessions removes sessions older than specified days.
func (s *Storage) CleanOldSessions(days int) int64 {
	query := fmt.Sprintf(`DELETE FROM sessions
		WHERE id IN (SELECT id FROM sessions WHERE created_at < now() - interval '%d days')`, days)

	result, err := s.db.Exec(query)
	if err != nil {
		return 0
	}

	n, _ := result.RowsAffected()
	return n
}
