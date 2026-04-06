// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"fmt"
	"log/slog"

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
func (s *Storage) RegisterWebPushSubscription(userID int64, request model.WebPushSubscription) error {
	query := `
		INSERT INTO webpush_subscriptions
			(user_id, endpoint, auth, p256dh, authscheme)
		VALUES
			($1, $2, $3, $4, $5)
		ON CONFLICT(endpoint) DO NOTHING
	`
	_, err := s.db.Exec(query, userID, request.Endpoint, request.Auth, request.Key, request.AuthScheme)

	return err
}

func (s *Storage) GetUserSubscriptions(userID int64) ([]model.WebPushSubscription, error) {
	query := `
		SELECT endpoint, auth, p256dh, authscheme
		FROM webpush_subscriptions
		WHERE user_id=$1
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch user webpush subscriptions: %v`, err)
	}
	defer rows.Close()

	var subscriptions []model.WebPushSubscription
	for rows.Next() {
		var subscription model.WebPushSubscription
		err := rows.Scan(
			&subscription.Endpoint,
			&subscription.Auth,
			&subscription.Key,
			&subscription.AuthScheme,
		)
		slog.Debug("Subscription found",
			slog.String("endpoint", subscription.Endpoint),
			slog.String("auth", subscription.Auth),
			slog.String("key", subscription.Key),
		)
		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch user webpush subscriptions: %v`, err)
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}
