// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

import (
	"encoding/xml"
	"fmt"
	"io"

	"miniflux.app/v2/internal/reader/encoding"
)

// parse reads an OPML file and returns a list of subscription.
func parse(data io.Reader) ([]subcription, error) {
	opmlDocument := &opmlDocument{}
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

func getSubscriptionsFromOutlines(outlines opmlOutlineCollection, category string) []subcription {
	subscriptions := make([]subcription, 0, len(outlines))

	for _, outline := range outlines {
		if outline.IsSubscription() {
			subscriptions = append(subscriptions, subcription{
				Title:        outline.GetTitle(),
				FeedURL:      outline.FeedURL,
				SiteURL:      outline.GetSiteURL(),
				Description:  outline.Description,
				CategoryName: category,
			})
		} else if outline.Outlines.HasChildren() {
			subscriptions = append(subscriptions, getSubscriptionsFromOutlines(outline.Outlines, outline.GetTitle())...)
		}
	}
	return subscriptions
}
