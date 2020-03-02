// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import (
	"time"

	"miniflux.app/crypto"
)

// APIKey represents an application API key.
type APIKey struct {
	ID          int64
	UserID      int64
	Token       string
	Description string
	LastUsedAt  *time.Time
	CreatedAt   time.Time
}

// NewAPIKey initializes a new APIKey.
func NewAPIKey(userID int64, description string) *APIKey {
	return &APIKey{
		UserID:      userID,
		Token:       crypto.GenerateRandomString(32),
		Description: description,
	}
}

// APIKeys represents a collection of API Key.
type APIKeys []*APIKey
