// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Storage) SetLastLogin(userID int64) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:SetLastLogin] userID=%d", userID))
	query := "UPDATE users SET last_login_at=now() WHERE id=$1"
	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("unable to update last login date: %v", err)
	}

	return nil
}

func (s *Storage) UserExists(username string) bool {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:UserExists] username=%s", username))

	var result int
	s.db.QueryRow(`SELECT count(*) as c FROM users WHERE username=$1`, username).Scan(&result)
	return result >= 1
}

func (s *Storage) AnotherUserExists(userID int64, username string) bool {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:AnotherUserExists] userID=%d, username=%s", userID, username))

	var result int
	s.db.QueryRow(`SELECT count(*) as c FROM users WHERE id != $1 AND username=$2`, userID, username).Scan(&result)
	return result >= 1
}

func (s *Storage) CreateUser(user *model.User) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:CreateUser] username=%s", user.Username))

	password, err := hashPassword(user.Password)
	if err != nil {
		return err
	}

	query := "INSERT INTO users (username, password, is_admin) VALUES ($1, $2, $3) RETURNING id"
	err = s.db.QueryRow(query, strings.ToLower(user.Username), password, user.IsAdmin).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("unable to create user: %v", err)
	}

	s.CreateCategory(&model.Category{Title: "All", UserID: user.ID})
	return nil
}

func (s *Storage) UpdateUser(user *model.User) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:UpdateUser] username=%s", user.Username))
	user.Username = strings.ToLower(user.Username)

	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return err
		}

		query := "UPDATE users SET username=$1, password=$2, is_admin=$3, theme=$4, language=$5, timezone=$6 WHERE id=$7"
		_, err = s.db.Exec(query, user.Username, hashedPassword, user.IsAdmin, user.Theme, user.Language, user.Timezone, user.ID)
		if err != nil {
			return fmt.Errorf("unable to update user: %v", err)
		}
	} else {
		query := "UPDATE users SET username=$1, is_admin=$2, theme=$3, language=$4, timezone=$5 WHERE id=$6"
		_, err := s.db.Exec(query, user.Username, user.IsAdmin, user.Theme, user.Language, user.Timezone, user.ID)
		if err != nil {
			return fmt.Errorf("unable to update user: %v", err)
		}
	}

	return nil
}

func (s *Storage) GetUserById(userID int64) (*model.User, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:GetUserById] userID=%d", userID))

	var user model.User
	row := s.db.QueryRow("SELECT id, username, is_admin, theme, language, timezone FROM users WHERE id = $1", userID)
	err := row.Scan(&user.ID, &user.Username, &user.IsAdmin, &user.Theme, &user.Language, &user.Timezone)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch user: %v", err)
	}

	return &user, nil
}

func (s *Storage) GetUserByUsername(username string) (*model.User, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:GetUserByUsername] username=%s", username))

	var user model.User
	row := s.db.QueryRow("SELECT id, username, is_admin, theme, language, timezone FROM users WHERE username=$1", username)
	err := row.Scan(&user.ID, &user.Username, &user.IsAdmin, &user.Theme, &user.Language, &user.Timezone)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch user: %v", err)
	}

	return &user, nil
}

func (s *Storage) RemoveUser(userID int64) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:RemoveUser] userID=%d", userID))

	result, err := s.db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("unable to remove this user: %v", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to remove this user: %v", err)
	}

	if count == 0 {
		return errors.New("nothing has been removed.")
	}

	return nil
}

func (s *Storage) GetUsers() (model.Users, error) {
	defer helper.ExecutionTime(time.Now(), "[Storage:GetUsers]")

	var users model.Users
	rows, err := s.db.Query("SELECT id, username, is_admin, theme, language, timezone, last_login_at FROM users ORDER BY username ASC")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch users: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.IsAdmin,
			&user.Theme,
			&user.Language,
			&user.Timezone,
			&user.LastLoginAt,
		)

		if err != nil {
			return nil, fmt.Errorf("unable to fetch users row: %v", err)
		}

		users = append(users, &user)
	}

	return users, nil
}

func (s *Storage) CheckPassword(username, password string) error {
	defer helper.ExecutionTime(time.Now(), "[Storage:CheckPassword]")

	var hash string
	username = strings.ToLower(username)

	err := s.db.QueryRow("SELECT password FROM users WHERE username=$1", username).Scan(&hash)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Unable to find this user: %s\n", username)
	} else if err != nil {
		return fmt.Errorf("Unable to fetch user: %v\n", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("Invalid password for %s\n", username)
	}

	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
