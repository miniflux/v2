// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml // import "miniflux.app/reader/opml"

import "encoding/xml"

type opml struct {
	XMLName  xml.Name  `xml:"opml"`
	Version  string    `xml:"version,attr"`
	Outlines []outline `xml:"body>outline"`
}

type outline struct {
	Title    string    `xml:"title,attr,omitempty"`
	Text     string    `xml:"text,attr"`
	FeedURL  string    `xml:"xmlUrl,attr,omitempty"`
	SiteURL  string    `xml:"htmlUrl,attr,omitempty"`
	Outlines []outline `xml:"outline,omitempty"`
}

func (o *outline) GetTitle() string {
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

func (o *outline) GetSiteURL() string {
	if o.SiteURL != "" {
		return o.SiteURL
	}

	return o.FeedURL
}

func (o *outline) IsCategory() bool {
	return o.Text != "" && o.SiteURL == "" && o.FeedURL == ""
}

func (o *outline) Append(subscriptions SubcriptionList, category string) SubcriptionList {
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

func (o *opml) Transform() SubcriptionList {
	var subscriptions SubcriptionList

	for _, outline := range o.Outlines {
		if outline.IsCategory() {
			for _, element := range outline.Outlines {
				subscriptions = element.Append(subscriptions, outline.Text)
			}
		} else {
			subscriptions = outline.Append(subscriptions, "")
		}
	}

	return subscriptions
}
