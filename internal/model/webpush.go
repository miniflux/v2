// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

// A request to subscribe to a new webpush
type WebPushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Key      string `json:"key"`
	Auth     string `json:"auth"`
}
