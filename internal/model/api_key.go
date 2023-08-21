// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"time"

	"miniflux.app/v2/internal/crypto"
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
