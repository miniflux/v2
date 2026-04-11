// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"miniflux.app/v2/internal/model"
)

// CreateWebSession persists a new web session built via model.NewWebSession.
func (s *Storage) CreateWebSession(session *model.WebSession) error {
	if session == nil {
		return errors.New(`store: web session is nil`)
	}

	stateJSON, err := session.MarshalState()
	if err != nil {
		return fmt.Errorf(`store: unable to serialize web session state: %v`, err)
	}

	query := `
		INSERT INTO web_sessions (
			id,
			secret_hash,
			user_agent,
			ip,
			state
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`

	err = s.db.QueryRow(
		query,
		session.ID,
		session.SecretHash,
		session.UserAgent,
		sql.NullString{String: session.IP, Valid: session.IP != ""},
		stateJSON,
	).Scan(&session.CreatedAt)
	if err != nil {
		return fmt.Errorf(`store: unable to create web session: %v`, err)
	}

	return nil
}

// WebSessionsByUserID returns web sessions for the given user.
func (s *Storage) WebSessionsByUserID(userID int64) ([]model.WebSession, error) {
	query := `
		SELECT
			id,
			secret_hash,
			user_id,
			created_at,
			user_agent,
			ip,
			state
		FROM
			web_sessions
		WHERE
			user_id=$1
		ORDER BY
			created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch web sessions: %v`, err)
	}
	defer rows.Close()

	var sessions []model.WebSession

	for rows.Next() {
		session, err := scanWebSession(rows)
		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch web session row: %v`, err)
		}

		sessions = append(sessions, *session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`store: unable to fetch web sessions: %v`, err)
	}

	return sessions, nil
}

// WebSessionByID returns the web session identified by id, or nil if not found.
func (s *Storage) WebSessionByID(sessionID string) (*model.WebSession, error) {
	if sessionID == "" {
		return nil, nil
	}

	row := s.db.QueryRow(`
		SELECT
			id,
			secret_hash,
			user_id,
			created_at,
			user_agent,
			ip,
			state
		FROM
			web_sessions
		WHERE
			id=$1
	`, sessionID)

	session, err := scanWebSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch web session: %v`, err)
	}

	return session, nil
}

// RotateWebSession persists a session whose identity has been rotated via
// (*model.WebSession).Rotate(), updating the row previously keyed by oldID.
func (s *Storage) RotateWebSession(oldID string, session *model.WebSession) error {
	if session == nil {
		return errors.New(`store: web session is nil`)
	}

	if oldID == "" || session.ID == "" {
		return errors.New(`store: web session ID cannot be empty`)
	}

	stateJSON, err := session.MarshalState()
	if err != nil {
		return fmt.Errorf(`store: unable to serialize web session state: %v`, err)
	}

	err = s.db.QueryRow(`
		UPDATE
			web_sessions
		SET
			id=$2,
			secret_hash=$3,
			user_id=$4,
			state=$5,
			created_at=now()
		WHERE
			id=$1
		RETURNING created_at
	`,
		oldID,
		session.ID,
		session.SecretHash,
		session.NullUserID(),
		stateJSON,
	).Scan(&session.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(`store: nothing has been updated`)
		}
		return fmt.Errorf(`store: unable to rotate web session: %v`, err)
	}

	return nil
}

// UpdateWebSession updates the mutable fields of a web session.
func (s *Storage) UpdateWebSession(session *model.WebSession) error {
	if session == nil {
		return errors.New(`store: web session is nil`)
	}

	if session.ID == "" {
		return errors.New(`store: web session ID cannot be empty`)
	}

	query := `
		UPDATE
			web_sessions
		SET
			user_id=$2,
			state=$3
		WHERE
			id=$1
	`

	stateJSON, err := session.MarshalState()
	if err != nil {
		return fmt.Errorf(`store: unable to serialize web session state: %v`, err)
	}

	result, err := s.db.Exec(
		query,
		session.ID,
		session.NullUserID(),
		stateJSON,
	)
	if err != nil {
		return fmt.Errorf(`store: unable to update web session: %v`, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(`store: unable to update web session: %v`, err)
	}

	if count != 1 {
		return errors.New(`store: nothing has been updated`)
	}

	return nil
}

// RemoveUserWebSession removes a web session for the given user if present.
func (s *Storage) RemoveUserWebSession(userID int64, sessionID string) error {
	if _, err := s.db.Exec(`DELETE FROM web_sessions WHERE user_id=$1 AND id=$2`, userID, sessionID); err != nil {
		return fmt.Errorf(`store: unable to remove this web session: %v`, err)
	}

	return nil
}

// CleanOldWebSessions removes web sessions older than the specified interval (24h minimum).
func (s *Storage) CleanOldWebSessions(interval time.Duration) (int64, error) {
	query := `
		DELETE FROM
			web_sessions
		WHERE
			created_at < now() - $1::interval
	`

	days := max(int(interval/(24*time.Hour)), 1)

	result, err := s.db.Exec(query, fmt.Sprintf("%d days", days))
	if err != nil {
		return 0, fmt.Errorf(`store: unable to clean old web sessions: %v`, err)
	}

	n, _ := result.RowsAffected()
	return n, nil
}

// FlushAllSessions removes all sessions from the database.
func (s *Storage) FlushAllSessions() error {
	if _, err := s.db.Exec(`DELETE FROM web_sessions`); err != nil {
		return fmt.Errorf(`store: unable to delete all web sessions: %v`, err)
	}
	return nil
}

type webSessionScanner interface {
	Scan(dest ...any) error
}

func scanWebSession(scanner webSessionScanner) (*model.WebSession, error) {
	var session model.WebSession
	var userID sql.NullInt64
	var ip sql.NullString
	var stateRaw []byte

	err := scanner.Scan(
		&session.ID,
		&session.SecretHash,
		&userID,
		&session.CreatedAt,
		&session.UserAgent,
		&ip,
		&stateRaw,
	)
	if err != nil {
		return nil, err
	}

	session.ScanUserID(userID)

	if ip.Valid {
		session.IP = ip.String
	}

	if err := session.UnmarshalState(stateRaw); err != nil {
		return nil, err
	}

	return &session, nil
}
