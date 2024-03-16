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
	// Version is the version of the RSS specification.
	Version string `xml:"rss version,attr"`

	// Channel is the main container for the RSS feed.
	Channel RSSChannel `xml:"rss channel"`
}

type RSSChannel struct {
	// Title is the name of the channel.
	Title string `xml:"rss title"`

	// Link is the URL to the HTML website corresponding to the channel.
	Link string `xml:"rss link"`

	// Description is a phrase or sentence describing the channel.
	Description string `xml:"rss description"`

	// Language is the language the channel is written in.
	// A list of allowable values for this element, as provided by Netscape, is here: https://www.rssboard.org/rss-language-codes.
	// You may also use values defined by the W3C: https://www.w3.org/TR/REC-html40/struct/dirlang.html#langcodes.
	Language string `xml:"rss language"`

	// Copyright is a string indicating the copyright.
	Copyright string `xml:"rss copyRight"`

	// ManagingEditor is the email address for the person responsible for editorial content.
	ManagingEditor string `xml:"rss managingEditor"`

	// Webmaster is the email address for the person responsible for technical issues relating to the channel.
	Webmaster string `xml:"rss webMaster"`

	// PubDate is the publication date for the content in the channel.
	// All date-times in RSS conform to the Date and Time Specification of RFC 822, with the exception that the year may be expressed with two characters or four characters (four preferred).
	PubDate string `xml:"rss pubDate"`

	// LastBuildDate is the last time the content of the channel changed.
	LastBuildDate string `xml:"rss lastBuildDate"`

	// Categories is a collection of categories to which the channel belongs.
	Categories []string `xml:"rss category"`

	// Generator is a string indicating the program used to generate the channel.
	Generator string `xml:"rss generator"`

	// Docs is a URL that points to the documentation for the format used in the RSS file.
	DocumentationURL string `xml:"rss docs"`

	// Cloud is a web service that supports the rssCloud interface which can be implemented in HTTP-POST, XML-RPC or SOAP 1.1.
	Cloud *RSSCloud `xml:"rss cloud"`

	// Image specifies a GIF, JPEG or PNG image that can be displayed with the channel.
	Image *RSSImage `xml:"rss image"`

	// TTL is a number of minutes that indicates how long a channel can be cached before refreshing from the source.
	TTL string `xml:"rss ttl"`

	// SkipHours is a hint for aggregators telling them which hours they can skip.
	// An XML element that contains up to 24 <hour> sub-elements whose value is a number between 0 and 23,
	// representing a time in GMT, when aggregators,
	// if they support the feature, may not read the channel on hours listed in the skipHours element.
	SkipHours []string `xml:"rss skipHours>hour"`

	// SkipDays is a hint for aggregators telling them which days they can skip.
	// An XML element that contains up to seven <day> sub-elements whose value is Monday, Tuesday, Wednesday, Thursday, Friday, Saturday or Sunday.
	SkipDays []string `xml:"rss skipDays>day"`

	// Items is a collection of items.
	Items []RSSItem `xml:"rss item"`

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
	// Title is the title of the item.
	Title string `xml:"rss title"`

	// Link is the URL of the item.
	Link string `xml:"rss link"`

	// Description is the item synopsis.
	Description string `xml:"rss description"`

	// Author is the email address of the author of the item.
	Author RSSAuthor `xml:"rss author"`

	// <category> is an optional sub-element of <item>.
	// It has one optional attribute, domain, a string that identifies a categorization taxonomy.
	Categories []string `xml:"rss category"`

	// <comments> is an optional sub-element of <item>.
	// If present, it contains the URL of the comments page for the item.
	CommentsURL string `xml:"rss comments"`

	// <enclosure> is an optional sub-element of <item>.
	// It has three required attributes. url says where the enclosure is located,
	// length says how big it is in bytes, and type says what its type is, a standard MIME type.
	Enclosures []RSSEnclosure `xml:"rss enclosure"`

	// <guid> is an optional sub-element of <item>.
	// It's a string that uniquely identifies the item.
	// When present, an aggregator may choose to use this string to determine if an item is new.
	//
	// There are no rules for the syntax of a guid.
	// Aggregators must view them as a string.
	// It's up to the source of the feed to establish the uniqueness of the string.
	//
	// If the guid element has an attribute named isPermaLink with a value of true,
	// the reader may assume that it is a permalink to the item, that is, a url that can be opened in a Web browser,
	// that points to the full item described by the <item> element.
	//
	// isPermaLink is optional, its default value is true.
	// If its value is false, the guid may not be assumed to be a url, or a url to anything in particular.
	GUID RSSGUID `xml:"rss guid"`

	// <pubDate> is the publication date of the item.
	// Its value is a string in RFC 822 format.
	PubDate string `xml:"rss pubDate"`

	// <source> is an optional sub-element of <item>.
	// Its value is the name of the RSS channel that the item came from, derived from its <title>.
	// It has one required attribute, url, which contains the URL of the RSS channel.
	Source RSSSource `xml:"rss source"`

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
