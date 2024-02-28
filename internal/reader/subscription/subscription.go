// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package subscription // import "miniflux.app/v2/internal/reader/subscription"

import "fmt"

// Subscription represents a feed subscription.
type Subscription struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

func NewSubscription(title, url, kind string) *Subscription {
	return &Subscription{Title: title, URL: url, Type: kind}
}

func (s Subscription) String() string {
	return fmt.Sprintf(`Title=%q, URL=%q, Type=%q`, s.Title, s.URL, s.Type)
}

// Subscriptions represents a list of subscription.
type Subscriptions []*Subscription
