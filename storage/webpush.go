// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"fmt"
	"miniflux.app/model"
	"time"

	"miniflux.app/timer"
)

// CreateWebpushSubscription creates a new WebpushSubscription.
func (s *Storage) CreateWebpushSubscription(subscription *model.WebpushSubscription) error {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:CreateWebpushSubscription] user_id=%d, subscription=%s", subscription.UserID, subscription.Subscription))

	query := `
		INSERT INTO webpush_subscriptions
		(user_id, subscription)
		VALUES
		($1, $2)
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		subscription.UserID,
		subscription.Subscription,
	).Scan(&subscription.ID)

	if err != nil {
		return fmt.Errorf("unable to create WebpushSubscription: %v", err)
	}

	return nil
}

func (s *Storage) GetSubscriptions(userID int64) (model.UserWebpushSubscriptions, error) {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:WebPushSubscriptions] userID=%d", userID))

	webpushSubscriptions := make(model.UserWebpushSubscriptions, 0)
	query := `SELECT
		id, user_id, subscription
		FROM webpush_subscriptions
		WHERE user_id=$1`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch WebPush subscriptions: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var webpush_subscription model.WebpushSubscription

		err := rows.Scan(
			&webpush_subscription.ID,
			&webpush_subscription.UserID,
			&webpush_subscription.Subscription)

		if err != nil {
			return nil, fmt.Errorf("unable to fetch WebPush subscriptions row: %v", err)
		}

		webpushSubscriptions = append(webpushSubscriptions, &webpush_subscription)
	}

	return webpushSubscriptions, nil
}
