// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package subscription

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
