// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml

type Subcription struct {
	Title        string
	SiteURL      string
	FeedURL      string
	CategoryName string
}

func (s Subcription) Equals(subscription *Subcription) bool {
	return s.Title == subscription.Title && s.SiteURL == subscription.SiteURL && s.FeedURL == subscription.FeedURL && s.CategoryName == subscription.CategoryName
}

type SubcriptionList []*Subcription
