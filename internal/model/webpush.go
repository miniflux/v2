// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

// A request to subscribe to a new webpush
type WebPushSubscription struct {
	Endpoint   string `json:"endpoint"`
	Key        string `json:"key"`
	Auth       string `json:"auth"`
	AuthScheme string `json:"authscheme"`
}

type Notification struct {
	FeedTitle    string `json:"feed_title"`
	EntryTitle   string `json:"entry_title"`
	EntryContent string `json:"entry_content"`
}
