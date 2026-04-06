// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package webpush // import "miniflux.app/v2/internal/reader/webpush"

import (
	"encoding/json"
	"log/slog"

	"miniflux.app/v2/internal/model"

	"github.com/SherClockHolmes/webpush-go"
)

func SendPush(subscriptions []model.WebPushSubscription, notification model.Notification, vapidPublicKey string, vapidPrivateKey string) error {
	slog.Debug(
		"Number of webpush subscriptions",
		slog.Int("length", len(subscriptions)),
	)

	for index := range subscriptions {

		var subs webpush.Subscription
		subs.Endpoint = subscriptions[index].Endpoint
		subs.Keys.Auth = subscriptions[index].Auth
		subs.Keys.P256dh = subscriptions[index].Key

		notificationJSON, err := json.Marshal(notification)
		if err != nil {
			return err
		}
		_, err = webpush.SendNotification([]byte(notificationJSON), &subs, &webpush.Options{
			// AuthScheme:      "vapid", // Not yet supported by the lib
			Subscriber:      "example@example.com", // Do not include "mailto:"
			VAPIDPublicKey:  vapidPublicKey,
			VAPIDPrivateKey: vapidPrivateKey,
			TTL:             30,
		})
		if err != nil {
			return err
		}
		slog.Debug("Sent WebPush notification")

	}
	return nil
}
