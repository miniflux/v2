// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package webpush // import "miniflux.app/v2/internal/reader/webpush"

import (
	"encoding/json"
	"log/slog"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"

	"github.com/SherClockHolmes/webpush-go"
)

func SendPush(subscriptions []model.WebPushSubscription, notification model.Notification, vapidPublicKey string, vapidPrivateKey string, userID int64, store *storage.Storage) error {
	for index := range subscriptions {
		var subs webpush.Subscription
		subs.Endpoint = subscriptions[index].Endpoint
		subs.Keys.Auth = subscriptions[index].Auth
		subs.Keys.P256dh = subscriptions[index].Key

		notificationJSON, err := json.Marshal(notification)
		if err != nil {
			return err
		}
		response, err := webpush.SendNotification([]byte(notificationJSON), &subs, &webpush.Options{
			// AuthScheme:      "vapid", // Not yet supported by the lib
			Subscriber:      config.Opts.AdminEmail(),
			VAPIDPublicKey:  vapidPublicKey,
			VAPIDPrivateKey: vapidPrivateKey,
			TTL:             30,
		})
		slog.Debug("Sent WebPush notification", slog.String("endpoint", subs.Endpoint))

		// If we get a 401, 404 or 410 return codes, that means that the
		// subscription is invalid and needs to be removed.
		if response.StatusCode == 401 || response.StatusCode == 404 || response.StatusCode == 410 || err != nil {
			store.RemoveUserSubscription(userID, subs.Endpoint)
			slog.Debug(
				"Removed webpush subscription",
				slog.String("endpoint", subs.Endpoint),
			)
		}
	}
	return nil
}
