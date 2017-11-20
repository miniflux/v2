// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rss

import (
	"encoding/xml"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/feed/date"
	"github.com/miniflux/miniflux2/reader/processor"
	"github.com/miniflux/miniflux2/reader/sanitizer"
)

type rssFeed struct {
	XMLName      xml.Name  `xml:"rss"`
	Version      string    `xml:"version,attr"`
	Title        string    `xml:"channel>title"`
	Links        []rssLink `xml:"channel>link"`
	Language     string    `xml:"channel>language"`
	Description  string    `xml:"channel>description"`
	PubDate      string    `xml:"channel>pubDate"`
	ItunesAuthor string    `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd channel>author"`
	Items        []rssItem `xml:"channel>item"`
}

type rssLink struct {
	XMLName xml.Name
	Data    string `xml:",chardata"`
	Href    string `xml:"href,attr"`
}

type rssItem struct {
	GUID              string         `xml:"guid"`
	Title             string         `xml:"title"`
	Link              string         `xml:"link"`
	OriginalLink      string         `xml:"http://rssnamespace.org/feedburner/ext/1.0 origLink"`
	Description       string         `xml:"description"`
	Content           string         `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	PubDate           string         `xml:"pubDate"`
	Date              string         `xml:"http://purl.org/dc/elements/1.1/ date"`
	Authors           []rssAuthor    `xml:"author"`
	Creator           string         `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Enclosures        []rssEnclosure `xml:"enclosure"`
	OrigEnclosureLink string         `xml:"http://rssnamespace.org/feedburner/ext/1.0 origEnclosureLink"`
}

type rssAuthor struct {
	XMLName xml.Name
	Data    string `xml:",chardata"`
	Name    string `xml:"name"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length string `xml:"length,attr"`
}

func (r *rssFeed) GetSiteURL() string {
	for _, elem := range r.Links {
		if elem.XMLName.Space == "" {
			return elem.Data
		}
	}

	return ""
}

func (r *rssFeed) GetFeedURL() string {
	for _, elem := range r.Links {
		if elem.XMLName.Space == "http://www.w3.org/2005/Atom" {
			return elem.Href
		}
	}

	return ""
}

func (r *rssFeed) Transform() *model.Feed {
	feed := new(model.Feed)
	feed.SiteURL = r.GetSiteURL()
	feed.FeedURL = r.GetFeedURL()
	feed.Title = sanitizer.StripTags(r.Title)

	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, item := range r.Items {
		entry := item.Transform()

		if entry.Author == "" && r.ItunesAuthor != "" {
			entry.Author = r.ItunesAuthor
		}
		entry.Author = sanitizer.StripTags(entry.Author)

		if entry.URL == "" {
			entry.URL = feed.SiteURL
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}
func (i *rssItem) GetDate() time.Time {
	value := i.PubDate
	if i.Date != "" {
		value = i.Date
	}

	if value != "" {
		result, err := date.Parse(value)
		if err != nil {
			log.Println(err)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func (i *rssItem) GetAuthor() string {
	for _, element := range i.Authors {
		if element.Name != "" {
			return element.Name
		}

		if element.Data != "" {
			return element.Data
		}
	}

	return i.Creator
}

func (i *rssItem) GetHash() string {
	for _, value := range []string{i.GUID, i.Link} {
		if value != "" {
			return helper.Hash(value)
		}
	}

	return ""
}

func (i *rssItem) GetContent() string {
	if i.Content != "" {
		return i.Content
	}

	return i.Description
}

func (i *rssItem) GetURL() string {
	if i.OriginalLink != "" {
		return i.OriginalLink
	}

	return i.Link
}

func (i *rssItem) GetEnclosures() model.EnclosureList {
	enclosures := make(model.EnclosureList, 0)

	for _, enclosure := range i.Enclosures {
		length, _ := strconv.Atoi(enclosure.Length)
		enclosureURL := enclosure.URL

		if i.OrigEnclosureLink != "" {
			filename := path.Base(i.OrigEnclosureLink)
			if strings.Contains(enclosureURL, filename) {
				enclosureURL = i.OrigEnclosureLink
			}
		}

		enclosures = append(enclosures, &model.Enclosure{
			URL:      enclosureURL,
			MimeType: enclosure.Type,
			Size:     length,
		})
	}

	return enclosures
}

func (i *rssItem) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.URL = i.GetURL()
	entry.Date = i.GetDate()
	entry.Author = i.GetAuthor()
	entry.Hash = i.GetHash()
	entry.Content = processor.ItemContentProcessor(entry.URL, i.GetContent())
	entry.Title = sanitizer.StripTags(strings.Trim(i.Title, " \n\t"))
	entry.Enclosures = i.GetEnclosures()

	if entry.Title == "" {
		entry.Title = entry.URL
	}

	return entry
}
