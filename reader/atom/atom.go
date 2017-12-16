// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package atom

import (
	"encoding/xml"
	"strconv"
	"strings"
	"time"

	"github.com/miniflux/miniflux/helper"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/reader/date"
	"github.com/miniflux/miniflux/url"
)

type atomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	ID      string      `xml:"id"`
	Title   string      `xml:"title"`
	Author  atomAuthor  `xml:"author"`
	Links   []atomLink  `xml:"link"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID         string         `xml:"id"`
	Title      string         `xml:"title"`
	Updated    string         `xml:"updated"`
	Links      []atomLink     `xml:"link"`
	Summary    string         `xml:"summary"`
	Content    atomContent    `xml:"content"`
	MediaGroup atomMediaGroup `xml:"http://search.yahoo.com/mrss/ group"`
	Author     atomAuthor     `xml:"author"`
}

type atomAuthor struct {
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

type atomLink struct {
	URL    string `xml:"href,attr"`
	Type   string `xml:"type,attr"`
	Rel    string `xml:"rel,attr"`
	Length string `xml:"length,attr"`
}

type atomContent struct {
	Type string `xml:"type,attr"`
	Data string `xml:",chardata"`
	XML  string `xml:",innerxml"`
}

type atomMediaGroup struct {
	Description string `xml:"http://search.yahoo.com/mrss/ description"`
}

func (a *atomFeed) Transform() *model.Feed {
	feed := new(model.Feed)
	feed.FeedURL = getRelationURL(a.Links, "self")
	feed.SiteURL = getURL(a.Links)
	feed.Title = strings.TrimSpace(a.Title)

	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, entry := range a.Entries {
		item := entry.Transform()
		entryURL, err := url.AbsoluteURL(feed.SiteURL, item.URL)
		if err == nil {
			item.URL = entryURL
		}

		if item.Author == "" {
			item.Author = getAuthor(a.Author)
		}

		if item.Title == "" {
			item.Title = item.URL
		}

		feed.Entries = append(feed.Entries, item)
	}

	return feed
}

func (a *atomEntry) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.URL = getURL(a.Links)
	entry.Date = getDate(a)
	entry.Author = getAuthor(a.Author)
	entry.Hash = getHash(a)
	entry.Content = getContent(a)
	entry.Title = strings.TrimSpace(a.Title)
	entry.Enclosures = getEnclosures(a)
	return entry
}

func getURL(links []atomLink) string {
	for _, link := range links {
		if strings.ToLower(link.Rel) == "alternate" {
			return strings.TrimSpace(link.URL)
		}

		if link.Rel == "" && link.Type == "" {
			return strings.TrimSpace(link.URL)
		}
	}

	return ""
}

func getRelationURL(links []atomLink, relation string) string {
	for _, link := range links {
		if strings.ToLower(link.Rel) == relation {
			return strings.TrimSpace(link.URL)
		}
	}

	return ""
}

func getDate(a *atomEntry) time.Time {
	if a.Updated != "" {
		result, err := date.Parse(a.Updated)
		if err != nil {
			logger.Error("atom: %v", err)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func getContent(a *atomEntry) string {
	if a.Content.Type == "html" || a.Content.Type == "text" {
		return a.Content.Data
	}

	if a.Content.Type == "xhtml" {
		return a.Content.XML
	}

	if a.Summary != "" {
		return a.Summary
	}

	if a.MediaGroup.Description != "" {
		return a.MediaGroup.Description
	}

	return ""
}

func getHash(a *atomEntry) string {
	for _, value := range []string{a.ID, getURL(a.Links)} {
		if value != "" {
			return helper.Hash(value)
		}
	}

	return ""
}

func getEnclosures(a *atomEntry) model.EnclosureList {
	enclosures := make(model.EnclosureList, 0)

	for _, link := range a.Links {
		if strings.ToLower(link.Rel) == "enclosure" {
			length, _ := strconv.Atoi(link.Length)
			enclosures = append(enclosures, &model.Enclosure{URL: link.URL, MimeType: link.Type, Size: length})
		}
	}

	return enclosures
}

func getAuthor(author atomAuthor) string {
	if author.Name != "" {
		return strings.TrimSpace(author.Name)
	}

	if author.Email != "" {
		return strings.TrimSpace(author.Email)
	}

	return ""
}
