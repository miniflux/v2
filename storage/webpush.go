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
