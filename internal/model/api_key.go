// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"time"
)

// APIKey represents an application API key.
type APIKey struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Token       string     `json:"token"`
	Description string     `json:"description"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// APIKeys represents a collection of API Key.
type APIKeys []*APIKey

// APIKeyCreationRequest represents the request to create a new API Key.
type APIKeyCreationRequest struct {
	Description string `json:"description"`
}
