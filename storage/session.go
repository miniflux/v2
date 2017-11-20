// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"fmt"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
)

func (s *Storage) GetSessions(userID int64) (model.Sessions, error) {
	query := `SELECT id, user_id, token, created_at, user_agent, ip FROM sessions WHERE user_id=$1 ORDER BY id DESC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch sessions: %v", err)
	}
	defer rows.Close()

	var sessions model.Sessions
	for rows.Next() {
		var session model.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.Token,
			&session.CreatedAt,
			&session.UserAgent,
			&session.IP,
		)

		if err != nil {
			return nil, fmt.Errorf("unable to fetch session row: %v", err)
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

func (s *Storage) CreateSession(username, userAgent, ip string) (sessionID string, err error) {
	var userID int64

	err = s.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("unable to fetch UserID: %v", err)
	}

	token := helper.GenerateRandomString(64)
	query := "INSERT INTO sessions (token, user_id, user_agent, ip) VALUES ($1, $2, $3, $4)"
	_, err = s.db.Exec(query, token, userID, userAgent, ip)
	if err != nil {
		return "", fmt.Errorf("unable to create session: %v", err)
	}

	s.SetLastLogin(userID)

	return token, nil
}

func (s *Storage) GetSessionByToken(token string) (*model.Session, error) {
	var session model.Session

	query := "SELECT id, user_id, token, created_at, user_agent, ip FROM sessions WHERE token = $1"
	err := s.db.QueryRow(query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.CreatedAt,
		&session.UserAgent,
		&session.IP,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", token)
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch session: %v", err)
	}

	return &session, nil
}

func (s *Storage) RemoveSessionByToken(userID int64, token string) error {
	result, err := s.db.Exec(`DELETE FROM sessions WHERE user_id=$1 AND token=$2`, userID, token)
	if err != nil {
		return fmt.Errorf("unable to remove this session: %v", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to remove this session: %v", err)
	}

	if count != 1 {
		return fmt.Errorf("nothing has been removed")
	}

	return nil
}

func (s *Storage) RemoveSessionByID(userID, sessionID int64) error {
	result, err := s.db.Exec(`DELETE FROM sessions WHERE user_id=$1 AND id=$2`, userID, sessionID)
	if err != nil {
		return fmt.Errorf("unable to remove this session: %v", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to remove this session: %v", err)
	}

	if count != 1 {
		return fmt.Errorf("nothing has been removed")
	}

	return nil
}

func (s *Storage) FlushAllSessions() (err error) {
	_, err = s.db.Exec(`delete from sessions`)
	return
}
