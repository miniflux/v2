// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rss

import (
	"encoding/xml"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/miniflux/miniflux/crypto"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/reader/date"
	"github.com/miniflux/miniflux/url"
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
	Rel     string `xml:"rel,attr"`
}

type rssItem struct {
	GUID              string         `xml:"guid"`
	Title             string         `xml:"title"`
	Links             []rssLink      `xml:"link"`
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
	for _, element := range r.Links {
		if element.XMLName.Space == "" {
			return strings.TrimSpace(element.Data)
		}
	}

	return ""
}

func (r *rssFeed) GetFeedURL() string {
	for _, element := range r.Links {
		if element.XMLName.Space == "http://www.w3.org/2005/Atom" {
			return strings.TrimSpace(element.Href)
		}
	}

	return ""
}

func (r *rssFeed) Transform() *model.Feed {
	feed := new(model.Feed)
	feed.SiteURL = r.GetSiteURL()
	feed.FeedURL = r.GetFeedURL()
	feed.Title = strings.TrimSpace(r.Title)

	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, item := range r.Items {
		entry := item.Transform()

		if entry.Author == "" && r.ItunesAuthor != "" {
			entry.Author = r.ItunesAuthor
		}
		entry.Author = strings.TrimSpace(entry.Author)

		if entry.URL == "" {
			entry.URL = feed.SiteURL
		} else {
			entryURL, err := url.AbsoluteURL(feed.SiteURL, entry.URL)
			if err == nil {
				entry.URL = entryURL
			}
		}

		if entry.Title == "" {
			entry.Title = entry.URL
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

func (r *rssItem) GetDate() time.Time {
	value := r.PubDate
	if r.Date != "" {
		value = r.Date
	}

	if value != "" {
		result, err := date.Parse(value)
		if err != nil {
			logger.Error("rss: %v", err)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func (r *rssItem) GetAuthor() string {
	for _, element := range r.Authors {
		if element.Name != "" {
			return element.Name
		}

		if element.Data != "" {
			return element.Data
		}
	}

	return r.Creator
}

func (r *rssItem) GetHash() string {
	for _, value := range []string{r.GUID, r.GetURL()} {
		if value != "" {
			return crypto.Hash(value)
		}
	}

	return ""
}

func (r *rssItem) GetContent() string {
	if r.Content != "" {
		return r.Content
	}

	return r.Description
}

func (r *rssItem) GetURL() string {
	if r.OriginalLink != "" {
		return r.OriginalLink
	}

	for _, link := range r.Links {
		if link.XMLName.Space == "http://www.w3.org/2005/Atom" && link.Href != "" && isValidLinkRelation(link.Rel) {
			return strings.TrimSpace(link.Href)
		}

		if link.Data != "" {
			return strings.TrimSpace(link.Data)
		}
	}

	return ""
}

func (r *rssItem) GetEnclosures() model.EnclosureList {
	enclosures := make(model.EnclosureList, 0)

	for _, enclosure := range r.Enclosures {
		length, _ := strconv.Atoi(enclosure.Length)
		enclosureURL := enclosure.URL

		if r.OrigEnclosureLink != "" {
			filename := path.Base(r.OrigEnclosureLink)
			if strings.Contains(enclosureURL, filename) {
				enclosureURL = r.OrigEnclosureLink
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

func (r *rssItem) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.URL = r.GetURL()
	entry.Date = r.GetDate()
	entry.Author = r.GetAuthor()
	entry.Hash = r.GetHash()
	entry.Content = r.GetContent()
	entry.Title = strings.TrimSpace(r.Title)
	entry.Enclosures = r.GetEnclosures()
	return entry
}

func isValidLinkRelation(rel string) bool {
	switch rel {
	case "", "alternate", "enclosure", "related", "self", "via":
		return true
	default:
		if strings.HasPrefix(rel, "http") {
			return true
		}
		return false
	}
}
