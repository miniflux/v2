// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"fmt"
	"strings"

	"miniflux.app/logger"
	"miniflux.app/model"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// CountUsers returns the total number of users.
func (s *Storage) CountUsers() int {
	var result int
	err := s.db.QueryRow(`SELECT count(*) FROM users`).Scan(&result)
	if err != nil {
		return 0
	}

	return result
}

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
	hashedPassword := ""
	if user.Password != "" {
		hashedPassword, err = hashPassword(user.Password)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO users
			(username, password, is_admin, google_id, openid_connect_id)
		VALUES
			(LOWER($1), $2, $3, $4, $5)
		RETURNING
			id,
			username,
			is_admin,
			language,
			theme,
			timezone,
			entry_direction,
			entries_per_page,
			keyboard_shortcuts,
			show_reading_time,
			entry_swipe,
			stylesheet,
			google_id,
			openid_connect_id
	`

	err = s.db.QueryRow(query, user.Username, hashedPassword, user.IsAdmin, user.GoogleID, user.OpenIDConnectID).Scan(
		&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.Language,
		&user.Theme,
		&user.Timezone,
		&user.EntryDirection,
		&user.EntriesPerPage,
		&user.KeyboardShortcuts,
		&user.ShowReadingTime,
		&user.EntrySwipe,
		&user.Stylesheet,
		&user.GoogleID,
		&user.OpenIDConnectID,
	)
	if err != nil {
		return fmt.Errorf(`store: unable to create user: %v`, err)
	}

	s.CreateCategory(&model.Category{Title: "All", UserID: user.ID})
	s.CreateIntegration(user.ID)
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
				entries_per_page=$8,
				keyboard_shortcuts=$9,
				show_reading_time=$10,
				entry_swipe=$11,
				stylesheet=$12,
				google_id=$13,
				openid_connect_id=$14
			WHERE
				id=$15
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
			user.EntriesPerPage,
			user.KeyboardShortcuts,
			user.ShowReadingTime,
			user.EntrySwipe,
			user.Stylesheet,
			user.GoogleID,
			user.OpenIDConnectID,
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
				entries_per_page=$7,
				keyboard_shortcuts=$8,
				show_reading_time=$9,
				entry_swipe=$10,
				stylesheet=$11,
				google_id=$12,
				openid_connect_id=$13
			WHERE
				id=$14
		`

		_, err := s.db.Exec(
			query,
			user.Username,
			user.IsAdmin,
			user.Theme,
			user.Language,
			user.Timezone,
			user.EntryDirection,
			user.EntriesPerPage,
			user.KeyboardShortcuts,
			user.ShowReadingTime,
			user.EntrySwipe,
			user.Stylesheet,
			user.GoogleID,
			user.OpenIDConnectID,
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
			entries_per_page,
			keyboard_shortcuts,
			show_reading_time,
			entry_swipe,
			last_login_at,
			stylesheet,
			google_id,
			openid_connect_id
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
			entries_per_page,
			keyboard_shortcuts,
			show_reading_time,
			entry_swipe,
			last_login_at,
			stylesheet,
			google_id,
			openid_connect_id
		FROM
			users
		WHERE
			username=LOWER($1)
	`
	return s.fetchUser(query, username)
}

// UserByField finds a user by a field value.
func (s *Storage) UserByField(field, value string) (*model.User, error) {
	query := `
		SELECT
			id,
			username,
			is_admin,
			theme,
			language,
			timezone,
			entry_direction,
			entries_per_page,
			keyboard_shortcuts,
			show_reading_time,
			entry_swipe,
			last_login_at,
			stylesheet,
			google_id,
			openid_connect_id
		FROM
			users
		WHERE
			%s=$1
	`
	return s.fetchUser(fmt.Sprintf(query, pq.QuoteIdentifier(field)), value)
}

// AnotherUserWithFieldExists returns true if a user has the value set for the given field.
func (s *Storage) AnotherUserWithFieldExists(userID int64, field, value string) bool {
	var result bool
	s.db.QueryRow(fmt.Sprintf(`SELECT true FROM users WHERE id <> $1 AND %s=$2`, pq.QuoteIdentifier(field)), userID, value).Scan(&result)
	return result
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
			u.entries_per_page,
			u.keyboard_shortcuts,
			u.show_reading_time,
			u.entry_swipe,
			u.last_login_at,
			u.stylesheet,
			u.google_id,
			u.openid_connect_id
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
	user := model.NewUser()
	err := s.db.QueryRow(query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.Theme,
		&user.Language,
		&user.Timezone,
		&user.EntryDirection,
		&user.EntriesPerPage,
		&user.KeyboardShortcuts,
		&user.ShowReadingTime,
		&user.EntrySwipe,
		&user.LastLoginAt,
		&user.Stylesheet,
		&user.GoogleID,
		&user.OpenIDConnectID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch user: %v`, err)
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

// RemoveUserAsync deletes user data without locking the database.
func (s *Storage) RemoveUserAsync(userID int64) {
	go func() {
		deleteUserFeeds(s.db, userID)
		s.db.Exec(`DELETE FROM users WHERE id=$1`, userID)
		s.db.Exec(`DELETE FROM integrations WHERE user_id=$1`, userID)
	}()
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
			entries_per_page,
			keyboard_shortcuts,
			show_reading_time,
			entry_swipe,
			last_login_at,
			stylesheet,
			google_id,
			openid_connect_id
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
		user := model.NewUser()
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.IsAdmin,
			&user.Theme,
			&user.Language,
			&user.Timezone,
			&user.EntryDirection,
			&user.EntriesPerPage,
			&user.KeyboardShortcuts,
			&user.ShowReadingTime,
			&user.EntrySwipe,
			&user.LastLoginAt,
			&user.Stylesheet,
			&user.GoogleID,
			&user.OpenIDConnectID,
		)

		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch users row: %v`, err)
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

func deleteUserFeeds(db *sql.DB, userID int64) {
	query := `SELECT id FROM feeds WHERE user_id=$1`
	rows, err := db.Query(query, userID)
	if err != nil {
		logger.Error(`store: unable to get user feeds: %v`, err)
		return
	}
	defer rows.Close()

	var feedIDs []int64
	for rows.Next() {
		var feedID int64
		rows.Scan(&feedID)
		feedIDs = append(feedIDs, feedID)
	}

	worker := func(jobs <-chan int64, results chan<- bool) {
		for feedID := range jobs {
			deleteUserEntries(db, userID, feedID)
			db.Exec(`DELETE FROM feeds WHERE id=$1`, feedID)
			results <- true
		}
	}

	const numWorkers = 3
	numJobs := len(feedIDs)
	jobs := make(chan int64, numJobs)
	results := make(chan bool, numJobs)

	for w := 0; w < numWorkers; w++ {
		go worker(jobs, results)
	}

	for j := 0; j < numJobs; j++ {
		jobs <- feedIDs[j]
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		<-results
	}
}

func deleteUserEntries(db *sql.DB, userID int64, feedID int64) {
	query := `SELECT id FROM entries WHERE user_id=$1 AND feed_id=$2`
	rows, err := db.Query(query, userID, feedID)
	if err != nil {
		logger.Error(`store: unable to get user feed entries: %v`, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var entryID int64
		rows.Scan(&entryID)
		deleteUserEnclosures(db, userID, entryID)
		db.Exec(`DELETE FROM entries WHERE id=$1`, entryID)
	}
}

func deleteUserEnclosures(db *sql.DB, userID int64, entryID int64) {
	query := `SELECT id FROM enclosures WHERE user_id=$1 AND entry_id=$2`
	rows, err := db.Query(query, userID, entryID)
	if err != nil {
		logger.Error(`store: unable to get user entry enclosures: %v`, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var enclosureID int64
		rows.Scan(&enclosureID)
		go func() {
			db.Exec(`DELETE FROM enclosures WHERE id=$1`, enclosureID)
		}()
	}
}
