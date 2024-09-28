// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

import (
	"encoding/xml"
	"fmt"
	"io"

	"miniflux.app/v2/internal/reader/encoding"
)

// Parse reads an OPML file and returns a SubcriptionList.
func Parse(data io.Reader) (SubcriptionList, error) {
	opmlDocument := NewOPMLDocument()
	decoder := xml.NewDecoder(data)
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false
	decoder.CharsetReader = encoding.CharsetReader

	err := decoder.Decode(opmlDocument)
	if err != nil {
		return nil, fmt.Errorf("opml: unable to parse document: %w", err)
	}

	return getSubscriptionsFromOutlines(opmlDocument.Outlines, ""), nil
}

func getSubscriptionsFromOutlines(outlines opmlOutlineCollection, category string) (subscriptions SubcriptionList) {
	// NOTE Using a map for set semantics to deduplicate subscriptions
	subscriptionsMap := make(map[string]*Subcription)
	for _, outline := range outlines {
		if outline.IsSubscription() {
			subscription, ok := subscriptionsMap[outline.FeedURL]
			if !ok || subscription == nil {
				// Do not overwrite existing entry
				subscription = &Subcription{
					Title:       outline.GetTitle(),
					FeedURL:     outline.FeedURL,
					SiteURL:     outline.GetSiteURL(),
					Description: outline.Description,
				}
				subscriptions = append(subscriptions, subscription)
				subscriptionsMap[outline.FeedURL] = subscription
			}
			if category != "" {
				subscription.CategoryNames = append(subscription.CategoryNames, category)
			}
		} else if outline.Outlines.HasChildren() {
			children := getSubscriptionsFromOutlines(outline.Outlines, outline.GetTitle())
			for _, childSubscription := range children {
				childFeedURL := childSubscription.FeedURL
				subscription, ok := subscriptionsMap[childFeedURL]
				if ok && subscription != nil {
					// Do not overwrite existing entry
					subscription.CategoryNames = append(subscription.CategoryNames, childSubscription.CategoryNames...)
				} else {
					subscriptions = append(subscriptions, childSubscription)
					subscriptionsMap[childFeedURL] = childSubscription
				}
			}
		}
	}
	return subscriptions
}
