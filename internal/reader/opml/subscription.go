// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

// TODO rename type to Subscription
// Subcription represents a feed that will be imported or exported.
type Subcription struct {
	Title         string
	SiteURL       string
	FeedURL       string
	CategoryNames CategoryNameList
	Description   string
}

type CategoryNameList []string

// Equals compares two category lists
func (c1s *CategoryNameList) Equals(c2s *CategoryNameList) bool {
	if len(*c1s) != len(*c2s) {
		return false
	}
	for _, c1 := range *c1s {
		found := false
		for _, c2 := range *c2s {
			if c1 == c2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Equals compare two subscriptions.
func (s Subcription) Equals(subscription *Subcription) bool {
	return s.Title == subscription.Title && s.SiteURL == subscription.SiteURL &&
		s.FeedURL == subscription.FeedURL && s.CategoryNames.Equals(&subscription.CategoryNames) &&
		s.Description == subscription.Description
}

// TODO rename type to SubscriptionList
// SubcriptionList is a list of subscriptions.
type SubcriptionList []*Subcription
