// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"encoding/xml"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/reader/dublincore"
	"miniflux.app/v2/internal/reader/googleplay"
	"miniflux.app/v2/internal/reader/itunes"
	"miniflux.app/v2/internal/reader/media"
)

// Specs: https://www.rssboard.org/rss-specification
type RSS struct {
	Version string     `xml:"rss version,attr"`
	Channel RSSChannel `xml:"rss channel"`
}

type RSSChannel struct {
	Title          string    `xml:"rss title"`
	Link           string    `xml:"rss link"`
	Description    string    `xml:"rss description"`
	Language       string    `xml:"rss language"`
	Copyright      string    `xml:"rss copyRight"`
	ManagingEditor string    `xml:"rss managingEditor"`
	Webmaster      string    `xml:"rss webMaster"`
	PubDate        string    `xml:"rss pubDate"`
	LastBuildDate  string    `xml:"rss lastBuildDate"`
	Categories     []string  `xml:"rss category"`
	Generator      string    `xml:"rss generator"`
	Docs           string    `xml:"rss docs"`
	Cloud          *RSSCloud `xml:"rss cloud"`
	Image          *RSSImage `xml:"rss image"`
	TTL            string    `xml:"rss ttl"`
	SkipHours      []string  `xml:"rss skipHours>hour"`
	SkipDays       []string  `xml:"rss skipDays>day"`
	Items          []RSSItem `xml:"rss item"`
	AtomLinks
	itunes.ItunesChannelElement
	googleplay.GooglePlayChannelElement
}

type RSSCloud struct {
	Domain            string `xml:"domain,attr"`
	Port              string `xml:"port,attr"`
	Path              string `xml:"path,attr"`
	RegisterProcedure string `xml:"registerProcedure,attr"`
	Protocol          string `xml:"protocol,attr"`
}

type RSSImage struct {
	// URL is the URL of a GIF, JPEG or PNG image that represents the channel.
	URL string `xml:"url"`

	// Title describes the image, it's used in the ALT attribute of the HTML <img> tag when the channel is rendered in HTML.
	Title string `xml:"title"`

	// Link is the URL of the site, when the channel is rendered, the image is a link to the site.
	Link string `xml:"link"`
}

type RSSItem struct {
	Title       string         `xml:"rss title"`
	Link        string         `xml:"rss link"`
	Description string         `xml:"rss description"`
	Author      RSSAuthor      `xml:"rss author"`
	Categories  []string       `xml:"rss category"`
	CommentsURL string         `xml:"rss comments"`
	Enclosures  []RSSEnclosure `xml:"rss enclosure"`
	GUID        RSSGUID        `xml:"rss guid"`
	PubDate     string         `xml:"rss pubDate"`
	Source      RSSSource      `xml:"rss source"`
	dublincore.DublinCoreItemElement
	FeedBurnerItemElement
	media.MediaItemElement
	AtomAuthor
	AtomLinks
	itunes.ItunesItemElement
	googleplay.GooglePlayItemElement
}

type RSSAuthor struct {
	XMLName xml.Name
	Data    string `xml:",chardata"`
	Inner   string `xml:",innerxml"`
}

type RSSEnclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length string `xml:"length,attr"`
}

func (enclosure *RSSEnclosure) Size() int64 {
	if strings.TrimSpace(enclosure.Length) == "" {
		return 0
	}
	size, _ := strconv.ParseInt(enclosure.Length, 10, 0)
	return size
}

type RSSGUID struct {
	Data        string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
}

type RSSSource struct {
	URL  string `xml:"url,attr"`
	Name string `xml:",chardata"`
}
