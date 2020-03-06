// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"fmt"
	"strings"

	"miniflux.app/model"

	"github.com/lib/pq/hstore"
	"golang.org/x/crypto/bcrypt"
)

// SetLastLogin updates the last login date of a user.
func (s *Storage) SetLastLogin(userID int64) error {
	query := `UPDATE users SET last_login_at=now() WHERE id=$1`
	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf(`store: unable to update last login date: %v`, err)
	}

	return nil
}

// UserExists checks if a user exists by using the given username.
func (s *Storage) UserExists(username string) bool {
	var result bool
	s.db.QueryRow(`SELECT true FROM users WHERE username=LOWER($1)`, username).Scan(&result)
	return result
}

// AnotherUserExists checks if another user exists with the given username.
func (s *Storage) AnotherUserExists(userID int64, username string) bool {
	var result bool
	s.db.QueryRow(`SELECT true FROM users WHERE id != $1 AND username=LOWER($2)`, userID, username).Scan(&result)
	return result
}

// CreateUser creates a new user.
func (s *Storage) CreateUser(user *model.User) (err error) {
	password := ""
	extra := hstore.Hstore{Map: make(map[string]sql.NullString)}

	if user.Password != "" {
		password, err = hashPassword(user.Password)
		if err != nil {
			return err
		}
	}

	if len(user.Extra) > 0 {
		for key, value := range user.Extra {
			extra.Map[key] = sql.NullString{String: value, Valid: true}
		}
	}

	query := `
		INSERT INTO users
			(username, password, is_admin, extra)
		VALUES
			(LOWER($1), $2, $3, $4)
		RETURNING
			id, username, is_admin, language, theme, timezone, entry_direction, keyboard_shortcuts
	`

	err = s.db.QueryRow(query, user.Username, password, user.IsAdmin, extra).Scan(
		&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.Language,
		&user.Theme,
		&user.Timezone,
		&user.EntryDirection,
		&user.KeyboardShortcuts,
	)
	if err != nil {
		return fmt.Errorf(`store: unable to create user: %v`, err)
	}

	s.CreateCategory(&model.Category{Title: "All", UserID: user.ID})
	s.CreateIntegration(user.ID)
	return nil
}

// UpdateExtraField updates an extra field of the given user.
func (s *Storage) UpdateExtraField(userID int64, field, value string) error {
	query := fmt.Sprintf(`UPDATE users SET extra = hstore('%s', $1) WHERE id=$2`, field)
	_, err := s.db.Exec(query, value, userID)
	if err != nil {
		return fmt.Errorf(`store: unable to update user extra field: %v`, err)
	}
	return nil
}

// RemoveExtraField deletes an extra field for the given user.
func (s *Storage) RemoveExtraField(userID int64, field string) error {
	query := `UPDATE users SET extra = delete(extra, $1) WHERE id=$2`
	_, err := s.db.Exec(query, field, userID)
	if err != nil {
		return fmt.Errorf(`store: unable to remove user extra field: %v`, err)
	}
	return nil
}

// UpdateUser updates a user.
func (s *Storage) UpdateUser(user *model.User) error {
	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return err
		}

		query := `
			UPDATE users SET
				username=LOWER($1),
				password=$2,
				is_admin=$3,
				theme=$4,
				language=$5,
				timezone=$6,
				entry_direction=$7,
				keyboard_shortcuts=$8
			WHERE
				id=$9
		`

		_, err = s.db.Exec(
			query,
			user.Username,
			hashedPassword,
			user.IsAdmin,
			user.Theme,
			user.Language,
			user.Timezone,
			user.EntryDirection,
			user.KeyboardShortcuts,
			user.ID,
		)
		if err != nil {
			return fmt.Errorf(`store: unable to update user: %v`, err)
		}
	} else {
		query := `
			UPDATE users SET
				username=LOWER($1),
				is_admin=$2,
				theme=$3,
				language=$4,
				timezone=$5,
				entry_direction=$6,
				keyboard_shortcuts=$7
			WHERE
				id=$8
		`

		_, err := s.db.Exec(
			query,
			user.Username,
			user.IsAdmin,
			user.Theme,
			user.Language,
			user.Timezone,
			user.EntryDirection,
			user.KeyboardShortcuts,
			user.ID,
		)

		if err != nil {
			return fmt.Errorf(`store: unable to update user: %v`, err)
		}
	}

	return nil
}

// UserLanguage returns the language of the given user.
func (s *Storage) UserLanguage(userID int64) (language string) {
	err := s.db.QueryRow(`SELECT language FROM users WHERE id = $1`, userID).Scan(&language)
	if err != nil {
		return "en_US"
	}

	return language
}

// UserByID finds a user by the ID.
func (s *Storage) UserByID(userID int64) (*model.User, error) {
	query := `
		SELECT
			id,
			username,
			is_admin,
			theme,
			language,
			timezone,
			entry_direction,
			keyboard_shortcuts,
			last_login_at,
			extra
		FROM
			users
		WHERE
			id = $1
	`
	return s.fetchUser(query, userID)
}

// UserByUsername finds a user by the username.
func (s *Storage) UserByUsername(username string) (*model.User, error) {
	query := `
		SELECT
			id,
			username,
			is_admin,
			theme,
			language,
			timezone,
			entry_direction,
			keyboard_shortcuts,
			last_login_at,
			extra
		FROM
			users
		WHERE
			username=LOWER($1)
	`
	return s.fetchUser(query, username)
}

// UserByExtraField finds a user by an extra field value.
func (s *Storage) UserByExtraField(field, value string) (*model.User, error) {
	query := `
		SELECT
			id,
			username,
			is_admin,
			theme,
			language,
			timezone,
			entry_direction,
			keyboard_shortcuts,
			last_login_at,
			extra
		FROM
			users
		WHERE
			extra->$1=$2
	`
	return s.fetchUser(query, field, value)
}

// UserByAPIKey returns a User from an API Key.
func (s *Storage) UserByAPIKey(token string) (*model.User, error) {
	query := `
		SELECT
			u.id,
			u.username,
			u.is_admin,
			u.theme,
			u.language,
			u.timezone,
			u.entry_direction,
			u.keyboard_shortcuts,
			u.last_login_at,
			u.extra
		FROM
			users u
		LEFT JOIN
			api_keys ON api_keys.user_id=u.id
		WHERE
			api_keys.token = $1
	`
	return s.fetchUser(query, token)
}

func (s *Storage) fetchUser(query string, args ...interface{}) (*model.User, error) {
	var extra hstore.Hstore

	user := model.NewUser()
	err := s.db.QueryRow(query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.Theme,
		&user.Language,
		&user.Timezone,
		&user.EntryDirection,
		&user.KeyboardShortcuts,
		&user.LastLoginAt,
		&extra,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch user: %v`, err)
	}

	for key, value := range extra.Map {
		if value.Valid {
			user.Extra[key] = value.String
		}
	}

	return user, nil
}

// RemoveUser deletes a user.
func (s *Storage) RemoveUser(userID int64) error {
	ts, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf(`store: unable to start transaction: %v`, err)
	}

	if _, err := ts.Exec(`DELETE FROM users WHERE id=$1`, userID); err != nil {
		ts.Rollback()
		return fmt.Errorf(`store: unable to remove user #%d: %v`, userID, err)
	}

	if _, err := ts.Exec(`DELETE FROM integrations WHERE user_id=$1`, userID); err != nil {
		ts.Rollback()
		return fmt.Errorf(`store: unable to remove integration settings for user #%d: %v`, userID, err)
	}

	if err := ts.Commit(); err != nil {
		return fmt.Errorf(`store: unable to commit transaction: %v`, err)
	}

	return nil
}

// Users returns all users.
func (s *Storage) Users() (model.Users, error) {
	query := `
		SELECT
			id,
			username,
			is_admin,
			theme,
			language,
			timezone,
			entry_direction,
			keyboard_shortcuts,
			last_login_at,
			extra
		FROM
			users
		ORDER BY username ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch users: %v`, err)
	}
	defer rows.Close()

	var users model.Users
	for rows.Next() {
		var extra hstore.Hstore
		user := model.NewUser()
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.IsAdmin,
			&user.Theme,
			&user.Language,
			&user.Timezone,
			&user.EntryDirection,
			&user.KeyboardShortcuts,
			&user.LastLoginAt,
			&extra,
		)

		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch users row: %v`, err)
		}

		for key, value := range extra.Map {
			if value.Valid {
				user.Extra[key] = value.String
			}
		}

		users = append(users, user)
	}

	return users, nil
}

// CheckPassword validate the hashed password.
func (s *Storage) CheckPassword(username, password string) error {
	var hash string
	username = strings.ToLower(username)

	err := s.db.QueryRow("SELECT password FROM users WHERE username=$1", username).Scan(&hash)
	if err == sql.ErrNoRows {
		return fmt.Errorf(`store: unable to find this user: %s`, username)
	} else if err != nil {
		return fmt.Errorf(`store: unable to fetch user: %v`, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf(`store: invalid password for "%s" (%v)`, username, err)
	}

	return nil
}

// HasPassword returns true if the given user has a password defined.
func (s *Storage) HasPassword(userID int64) (bool, error) {
	var result bool
	query := `SELECT true FROM users WHERE id=$1 AND password <> ''`

	err := s.db.QueryRow(query, userID).Scan(&result)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf(`store: unable to execute query: %v`, err)
	}

	if result {
		return true, nil
	}
	return false, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
