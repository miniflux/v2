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

				ScraperRules:                outline.ScraperRules,
				RewriteRules:                outline.RewriteRules,
				UrlRewriteRules:             outline.UrlRewriteRules,
				BlocklistRules:              outline.BlocklistRules,
				KeeplistRules:               outline.KeeplistRules,
				BlockFilterEntryRules:       outline.BlockFilterEntryRules,
				KeepFilterEntryRules:        outline.KeepFilterEntryRules,
				UserAgent:                   outline.UserAgent,
				Cookie:                      outline.Cookie,
				ProxyURL:                    outline.ProxyURL,
				Crawler:                     outline.Crawler,
				IgnoreHTTPCache:             outline.IgnoreHTTPCache,
				FetchViaProxy:               outline.FetchViaProxy,
				Disabled:                    outline.Disabled,
				NoMediaPlayer:               outline.NoMediaPlayer,
				HideGlobally:                outline.HideGlobally,
				AllowSelfSignedCertificates: outline.AllowSelfSignedCertificates,
				DisableHTTP2:                outline.DisableHTTP2,
				IgnoreEntryUpdates:          outline.IgnoreEntryUpdates,
			})
		} else if outline.Outlines.HasChildren() {
			subscriptions = append(subscriptions, getSubscriptionsFromOutlines(outline.Outlines, outline.GetTitle())...)
		}
	}
	return subscriptions
}
