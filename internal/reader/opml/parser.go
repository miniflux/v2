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
	for _, outline := range outlines {
		if outline.IsSubscription() {
			subscriptions = append(subscriptions, &Subcription{
				Title:        outline.GetTitle(),
				FeedURL:      outline.FeedURL,
				SiteURL:      outline.GetSiteURL(),
				CategoryName: category,
			})
		} else if outline.Outlines.HasChildren() {
			subscriptions = append(subscriptions, getSubscriptionsFromOutlines(outline.Outlines, outline.GetTitle())...)
		}
	}
	return subscriptions
}
