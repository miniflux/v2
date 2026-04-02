// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"miniflux.app/v2/internal/model"
)

// Get the VAPID keys from the database
func (s *Storage) GetVAPIDKeys() (string, string, error) {

	var privateKey string
	var publicKey string
	query := `SELECT private_key, public_key FROM vapid_key`

	err := s.db.QueryRow(query).Scan(&privateKey, &publicKey)

	return privateKey, publicKey, err
}

// Register a new subscription
func (s *Storage) RegisterWebPushSubscription(userID int64, request model.WebPushSubscriptionRequest) error {
	query := `
		INSERT INTO webpush_subscriptions
			(user_id, endpoint, auth, p256dh)
		VALUES
			($1, $2, $3, $4)
	`
	_, err := s.db.Exec(query, userID, request.Endpoint, request.Auth, request.Key)

	return err
}
