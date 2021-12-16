// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml // import "miniflux.app/reader/opml"

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"sort"
	"time"

	"miniflux.app/logger"
)

// Serialize returns a SubcriptionList in OPML format.
func Serialize(subscriptions SubcriptionList) string {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	writer.WriteString(xml.Header)

	opmlDocument := convertSubscriptionsToOPML(subscriptions)
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "    ")
	if err := encoder.Encode(opmlDocument); err != nil {
		logger.Error("[OPML:Serialize] %v", err)
		return ""
	}

	return b.String()
}

func convertSubscriptionsToOPML(subscriptions SubcriptionList) *opmlDocument {
	opmlDocument := NewOPMLDocument()
	opmlDocument.Version = "2.0"
	opmlDocument.Header.Title = "Miniflux"
	opmlDocument.Header.DateCreated = time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST")

	groupedSubs := groupSubscriptionsByFeed(subscriptions)
	var categories []string
	for k := range groupedSubs {
		categories = append(categories, k)
	}
	sort.Strings(categories)

	for _, categoryName := range categories {
		category := opmlOutline{Text: categoryName}
		for _, subscription := range groupedSubs[categoryName] {
			category.Outlines = append(category.Outlines, opmlOutline{
				Title:   subscription.Title,
				Text:    subscription.Title,
				FeedURL: subscription.FeedURL,
				SiteURL: subscription.SiteURL,
			})
		}

		opmlDocument.Outlines = append(opmlDocument.Outlines, category)
	}

	return opmlDocument
}

func groupSubscriptionsByFeed(subscriptions SubcriptionList) map[string]SubcriptionList {
	groups := make(map[string]SubcriptionList)

	for _, subscription := range subscriptions {
		groups[subscription.CategoryName] = append(groups[subscription.CategoryName], subscription)
	}

	return groups
}
