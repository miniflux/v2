// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"log"
)

func Serialize(subscriptions SubcriptionList) string {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	writer.WriteString(xml.Header)

	opml := new(Opml)
	opml.Version = "2.0"
	for categoryName, subs := range groupSubscriptionsByFeed(subscriptions) {
		outline := Outline{Text: categoryName}

		for _, subscription := range subs {
			outline.Outlines = append(outline.Outlines, Outline{
				Title:   subscription.Title,
				Text:    subscription.Title,
				FeedURL: subscription.FeedURL,
				SiteURL: subscription.SiteURL,
			})
		}

		opml.Outlines = append(opml.Outlines, outline)
	}

	encoder := xml.NewEncoder(writer)
	encoder.Indent("  ", "    ")
	if err := encoder.Encode(opml); err != nil {
		log.Println(err)
		return ""
	}

	return b.String()
}

func groupSubscriptionsByFeed(subscriptions SubcriptionList) map[string]SubcriptionList {
	groups := make(map[string]SubcriptionList)

	for _, subscription := range subscriptions {
		// if subs, ok := groups[subscription.CategoryName]; !ok {
		// groups[subscription.CategoryName] = SubcriptionList{}
		// }

		groups[subscription.CategoryName] = append(groups[subscription.CategoryName], subscription)
	}

	return groups
}
