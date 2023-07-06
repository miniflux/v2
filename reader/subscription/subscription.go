// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package subscription // import "miniflux.app/reader/subscription"

import "fmt"

// Subscription represents a feed subscription.
type Subscription struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

func (s Subscription) String() string {
	return fmt.Sprintf(`Title="%s", URL="%s", Type="%s"`, s.Title, s.URL, s.Type)
}

// Subscriptions represents a list of subscription.
type Subscriptions []*Subscription
