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

// CreateToken creates a new token.
func (s *Storage) CreateToken() (*model.Token, error) {
	token := model.Token{
		ID:    helper.GenerateRandomString(32),
		Value: helper.GenerateRandomString(64),
	}

	query := "INSERT INTO tokens (id, value) VALUES ($1, $2)"
	_, err := s.db.Exec(query, token.ID, token.Value)
	if err != nil {
		return nil, fmt.Errorf("unable to create token: %v", err)
	}

	return &token, nil
}

// Token returns a Token.
func (s *Storage) Token(id string) (*model.Token, error) {
	var token model.Token

	query := "SELECT id, value FROM tokens WHERE id=$1"
	err := s.db.QueryRow(query, id).Scan(
		&token.ID,
		&token.Value,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found: %s", id)
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch token: %v", err)
	}

	return &token, nil
}
