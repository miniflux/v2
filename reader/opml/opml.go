// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml

import "encoding/xml"

type Opml struct {
	XMLName  xml.Name  `xml:"opml"`
	Version  string    `xml:"version,attr"`
	Outlines []Outline `xml:"body>outline"`
}

type Outline struct {
	Title    string    `xml:"title,attr,omitempty"`
	Text     string    `xml:"text,attr"`
	FeedURL  string    `xml:"xmlUrl,attr,omitempty"`
	SiteURL  string    `xml:"htmlUrl,attr,omitempty"`
	Outlines []Outline `xml:"outline,omitempty"`
}

func (o *Outline) GetTitle() string {
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

func (o *Outline) GetSiteURL() string {
	if o.SiteURL != "" {
		return o.SiteURL
	}

	return o.FeedURL
}

func (o *Outline) IsCategory() bool {
	return o.Text != "" && o.SiteURL == "" && o.FeedURL == ""
}

func (o *Outline) Append(subscriptions SubcriptionList, category string) SubcriptionList {
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

func (o *Opml) Transform() SubcriptionList {
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
