// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

// Subcription represents a feed that will be imported or exported.
type Subcription struct {
	Title        string
	SiteURL      string
	FeedURL      string
	CategoryName string
}

// Equals compare two subscriptions.
func (s Subcription) Equals(subscription *Subcription) bool {
	return s.Title == subscription.Title && s.SiteURL == subscription.SiteURL &&
		s.FeedURL == subscription.FeedURL && s.CategoryName == subscription.CategoryName
}

// SubcriptionList is a list of subscriptions.
type SubcriptionList []*Subcription
