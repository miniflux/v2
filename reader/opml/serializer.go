// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml // import "miniflux.app/reader/opml"

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"sort"

	"miniflux.app/logger"
)

// Serialize returns a SubcriptionList in OPML format.
func Serialize(subscriptions SubcriptionList) string {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	writer.WriteString(xml.Header)

	feeds := normalizeFeeds(subscriptions)
	encoder := xml.NewEncoder(writer)
	encoder.Indent("    ", "    ")
	if err := encoder.Encode(feeds); err != nil {
		logger.Error("[OPML:Serialize] %v", err)
		return ""
	}

	return b.String()
}

func groupSubscriptionsByFeed(subscriptions SubcriptionList) map[string]SubcriptionList {
	groups := make(map[string]SubcriptionList)

	for _, subscription := range subscriptions {
		groups[subscription.CategoryName] = append(groups[subscription.CategoryName], subscription)
	}

	return groups
}

func normalizeFeeds(subscriptions SubcriptionList) *opml {
	feeds := new(opml)
	feeds.Version = "2.0"

	groupedSubs := groupSubscriptionsByFeed(subscriptions)
	var categories []string
	for k := range groupedSubs {
		categories = append(categories, k)
	}
	sort.Strings(categories)

	for _, categoryName := range categories {
		category := outline{Text: categoryName}
		for _, subscription := range groupedSubs[categoryName] {
			category.Outlines = append(category.Outlines, outline{
				Title:   subscription.Title,
				Text:    subscription.Title,
				FeedURL: subscription.FeedURL,
				SiteURL: subscription.SiteURL,
			})
		}

		feeds.Outlines = append(feeds.Outlines, category)
	}

	return feeds
}
