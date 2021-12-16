// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml // import "miniflux.app/reader/opml"

import (
	"encoding/xml"
)

// Specs: http://opml.org/spec2.opml
type opmlDocument struct {
	XMLName  xml.Name      `xml:"opml"`
	Version  string        `xml:"version,attr"`
	Header   opmlHeader    `xml:"head"`
	Outlines []opmlOutline `xml:"body>outline"`
}

func NewOPMLDocument() *opmlDocument {
	return &opmlDocument{}
}

func (o *opmlDocument) GetSubscriptionList() SubcriptionList {
	var subscriptions SubcriptionList
	for _, outline := range o.Outlines {
		if len(outline.Outlines) > 0 {
			for _, element := range outline.Outlines {
				// outline.Text is only available in OPML v2.
				subscriptions = element.Append(subscriptions, outline.Text)
			}
		} else {
			subscriptions = outline.Append(subscriptions, "")
		}
	}

	return subscriptions
}

type opmlHeader struct {
	Title       string `xml:"title,omitempty"`
	DateCreated string `xml:"dateCreated,omitempty"`
	OwnerName   string `xml:"ownerName,omitempty"`
}

type opmlOutline struct {
	Title    string        `xml:"title,attr,omitempty"`
	Text     string        `xml:"text,attr"`
	FeedURL  string        `xml:"xmlUrl,attr,omitempty"`
	SiteURL  string        `xml:"htmlUrl,attr,omitempty"`
	Outlines []opmlOutline `xml:"outline,omitempty"`
}

func (o *opmlOutline) GetTitle() string {
	if o.Title != "" {
		return o.Title
	}

	if o.Text != "" {
		return o.Text
	}

	if o.SiteURL != "" {
		return o.SiteURL
	}

	if o.FeedURL != "" {
		return o.FeedURL
	}

	return ""
}

func (o *opmlOutline) GetSiteURL() string {
	if o.SiteURL != "" {
		return o.SiteURL
	}

	return o.FeedURL
}

func (o *opmlOutline) Append(subscriptions SubcriptionList, category string) SubcriptionList {
	if o.FeedURL != "" {
		subscriptions = append(subscriptions, &Subcription{
			Title:        o.GetTitle(),
			FeedURL:      o.FeedURL,
			SiteURL:      o.GetSiteURL(),
			CategoryName: category,
		})
	}

	return subscriptions
}
